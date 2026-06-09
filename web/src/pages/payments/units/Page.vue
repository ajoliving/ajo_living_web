<!--
 * 支付單元頁。
 * 1. 依 POS 大廈資料選擇繳費大廈、樓層與單位。
 * 2. 儲存到會員資料，供支付中心查詢賬單、訂單與歷史。
-->
<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import axios from 'axios';

import { useFeedbackStore } from '@/app/stores/feedback';
import { useSessionStore } from '@/app/stores/session';
import { fetchPosBuildings, fetchPosBuildingUnits } from '@/domains/building/api';
import type { PosBuilding, PosBuildingUnit } from '@/domains/building/model';
import { updateMe } from '@/domains/account/api';
import AppGlassSelect from '@/shared/components/base/AppGlassSelect.vue';
import BaseButton from '@/shared/components/base/BaseButton.vue';

interface PaymentUnitOption {
  label: string;
  value: string;
}

const feedbackStore = useFeedbackStore();
const sessionStore = useSessionStore();
const buildings = ref<PosBuilding[]>([]);
const units = ref<PosBuildingUnit[]>([]);
const selectedBuildingID = ref('');
const selectedFloor = ref('');
const selectedUnitID = ref('');
const isLoading = ref(false);
const isUnitsLoading = ref(false);
const isSaving = ref(false);
let latestUnitsRequestID = 0;
let isSyncingProfileSelection = false;
const unassignedFloor = '未指定樓層';

// 1. 讀取 API 錯誤訊息
const readErrorMessage = (error: unknown, fallback: string): string =>
  axios.isAxiosError<{ message?: string }>(error)
    ? error.response?.data?.message ?? fallback
    : fallback;

// 1.1 判斷是否為系統佔位電話帳號
const isSystemPhonePlaceholder = (countryCode?: string): boolean =>
  ['email', 'ismart'].includes((countryCode ?? '').trim().toLowerCase());

// 2. 讀取 POS 大廈 ID
const getBuildingID = (item: PosBuilding): string =>
  String(item.building_id ?? item.id ?? '').trim();

// 3. 讀取 POS 大廈名稱
const getBuildingName = (item: PosBuilding): string =>
  String(item.buildname_chi ?? item.buildname ?? item.name ?? getBuildingID(item)).trim();

// 4. 讀取 POS 單位 ID
const getUnitID = (item: PosBuildingUnit): string =>
  String(item.unit_id ?? item.id ?? '').trim();

// 5. 讀取 POS 單位樓層
const getUnitFloor = (item: PosBuildingUnit): string =>
  String(item.floor ?? '').trim();

// 5.1 讀取 POS 單位樓層顯示值
const getDisplayUnitFloor = (item: PosBuildingUnit): string =>
  getUnitFloor(item) || unassignedFloor;

// 6. 讀取 POS 單位名稱
const getUnitName = (item: PosBuildingUnit): string =>
  String(item.unit ?? item.unit_name ?? item.name ?? '').trim();

// 7. 取出數字字串
const digitsOnly = (value: unknown): string =>
  String(value ?? '').replace(/\D/g, '');

// 7.1 正規化 POS 權限 ID 列表
const normalizeTextList = (values: string[] | undefined): string[] => {
  const result: string[] = [];
  values?.forEach((item) => {
    const value = String(item ?? '').trim();
    if (value && !result.includes(value)) {
      result.push(value);
    }
  });
  return result;
};

// 8. 排除 POS 回傳的樓宇佔位資料
const isSelectableUnit = (buildingID: string, item: PosBuildingUnit): boolean => {
  const normalizedBuildingID = digitsOnly(buildingID).slice(0, 7);
  const normalizedUnitID = digitsOnly(getUnitID(item));
  const hasFloor = getUnitFloor(item).length > 0;
  const hasUnit = getUnitName(item).length > 0;

  return !(normalizedBuildingID && normalizedUnitID === normalizedBuildingID && !hasFloor && !hasUnit);
};

// 9. 按 POS 顯示規則排序
const compareDisplayCodes = (left: string, right: string): number =>
  left.localeCompare(right, 'en', {
    numeric: true,
    sensitivity: 'base',
  });

// 9.1 判斷 POS Staff 身份
const isPOSStaff = computed(() =>
  Boolean(sessionStore.me?.is_staff || sessionStore.me?.ismart_msg?.is_staff),
);

// 9.2 取得 POS 權限大廈
const allowedBuildingIDs = computed(() => {
  const message = sessionStore.me?.ismart_msg;
  if (message) {
    if (isPOSStaff.value) {
      const staffBuildings = normalizeTextList(message.staff_building_permissions);
      return staffBuildings.length > 0 ? staffBuildings : normalizeTextList(message.building);
    }

    return normalizeTextList(message.client_building_permissions);
  }

  return normalizeTextList(sessionStore.me?.bound_building_ids);
});

// 9.3 取得 POS 權限單位
const allowedUnitIDs = computed(() =>
  isPOSStaff.value
    ? []
    : normalizeTextList(
      sessionStore.me?.ismart_msg?.client_building_flat_units_permissions ?? sessionStore.me?.bound_flat_unit_ids,
    ),
);

// 9.4 取得目前可見大廈
const visibleBuildings = computed(() => {
  const allowed = new Set(allowedBuildingIDs.value);
  return buildings.value.filter((item) => allowed.has(getBuildingID(item)));
});

// 9.5 取得目前可見單位
const visibleUnits = computed(() => {
  if (isPOSStaff.value) {
    return units.value;
  }

  const allowed = new Set(allowedUnitIDs.value);
  return units.value.filter((item) => allowed.has(getUnitID(item)));
});

// 10. 建立大廈選項
const buildingOptions = computed<PaymentUnitOption[]>(() => {
  const options = visibleBuildings.value
    .slice()
    .sort((left, right) => getBuildingName(left).localeCompare(getBuildingName(right), 'en', {
      numeric: true,
      sensitivity: 'base',
    }))
    .map((item) => ({
      label: getBuildingName(item),
      value: getBuildingID(item),
    }))
    .filter((item) => item.value.length > 0);

  return [
    {
      label: isLoading.value ? '載入中' : '請選擇大廈',
      value: '',
    },
    ...options,
  ];
});

// 11. 建立樓層選項
const floorOptions = computed<PaymentUnitOption[]>(() => {
  const floors = Array.from(
    new Set(visibleUnits.value.map((item) => getDisplayUnitFloor(item)).filter(Boolean)),
  ).sort(compareDisplayCodes);

  return [
    {
      label: isUnitsLoading.value ? '載入中' : '請選擇樓層',
      value: '',
    },
    ...floors.map((floor) => ({
      label: floor,
      value: floor,
    })),
  ];
});

// 12. 建立單位選項
const unitOptions = computed<PaymentUnitOption[]>(() => {
  const unitItems = visibleUnits.value
    .filter((item) => getDisplayUnitFloor(item) === selectedFloor.value)
    .slice()
    .sort((left, right) => compareDisplayCodes(getUnitName(left), getUnitName(right)))
    .map((item) => {
      const unitID = getUnitID(item);
      const unitName = getUnitName(item);
      return {
        label: unitName,
        value: unitID,
      };
    })
    .filter((item) => item.label.length > 0 && item.value.length > 0);

  return [
    {
      label: isUnitsLoading.value ? '載入中' : '請選擇單位',
      value: '',
    },
    ...unitItems,
  ];
});

const selectedBuildingName = computed(() =>
  buildingOptions.value.find((item) => item.value === selectedBuildingID.value)?.label ?? '',
);
const selectedUnit = computed(() =>
  visibleUnits.value.find((item) => getUnitID(item) === selectedUnitID.value),
);
const selectedUnitName = computed(() =>
  selectedUnit.value
    ? getUnitName(selectedUnit.value)
    : '',
);
const currentLabel = computed(() => {
  const buildingName = selectedBuildingName.value;
  const unitName = selectedUnitName.value;
  return [buildingName, selectedFloor.value, unitName].filter(Boolean).join(' / ') || '尚未選擇單位';
});
const canSave = computed(() =>
  selectedBuildingID.value.length > 0 &&
    selectedFloor.value.length > 0 &&
    selectedUnitID.value.length > 0 &&
    Boolean(selectedUnit.value),
);

// 13. 同步目前會員單位
const syncFromProfile = (): void => {
  isSyncingProfileSelection = true;
  const profileBuildingID = sessionStore.me?.primary_community?.public_id?.trim() ?? '';
  selectedBuildingID.value = allowedBuildingIDs.value.includes(profileBuildingID)
    ? profileBuildingID
    : allowedBuildingIDs.value[0] ?? '';
  selectedFloor.value = sessionStore.me?.residence_floor || '';
  selectedUnitID.value = '';
};

// 13.1 按已保存樓層與單位名稱同步目前單位
const syncSelectedUnitFromProfile = (): void => {
  const floor = sessionStore.me?.residence_floor?.trim() ?? '';
  const unitName = sessionStore.me?.residence_unit?.trim() ?? '';
  if (!floor || !unitName) {
    selectedUnitID.value = '';
    return;
  }

  const matchedUnit = visibleUnits.value.find((item) =>
    getDisplayUnitFloor(item) === floor && getUnitName(item) === unitName,
  );
  selectedUnitID.value = matchedUnit ? getUnitID(matchedUnit) : '';
};

// 14. 載入大廈清單
const loadBuildings = async (): Promise<void> => {
  isLoading.value = true;
  try {
    buildings.value = await fetchPosBuildings();
  } catch (error) {
    buildings.value = [];
    feedbackStore.pushToast(readErrorMessage(error, '大廈資料載入失敗。'), 'error');
  } finally {
    isLoading.value = false;
  }
};

// 15. 載入指定大廈單位清單
const loadUnits = async (buildingID: string): Promise<void> => {
  const requestID = ++latestUnitsRequestID;
  const value = buildingID.trim();
  if (!value) {
    units.value = [];
    return;
  }

  isUnitsLoading.value = true;
  try {
    const result = await fetchPosBuildingUnits(value);
    if (requestID !== latestUnitsRequestID) {
      return;
    }
    units.value = result.filter((item) => isSelectableUnit(value, item));
    if (isSyncingProfileSelection) {
      syncSelectedUnitFromProfile();
      return;
    }
    if (selectedUnitID.value && !visibleUnits.value.some((item) => getUnitID(item) === selectedUnitID.value)) {
      selectedUnitID.value = '';
    }
  } catch (error) {
    if (requestID !== latestUnitsRequestID) {
      return;
    }
    units.value = [];
    feedbackStore.pushToast(readErrorMessage(error, '單位資料載入失敗。'), 'error');
  } finally {
    if (requestID === latestUnitsRequestID) {
      isUnitsLoading.value = false;
    }
  }
};

// 16. 儲存目前繳費單位
const handleSaveUnit = async (): Promise<void> => {
  if (!canSave.value) {
    feedbackStore.pushToast('請先選擇大廈、樓層與單位。', 'error');
    return;
  }

  isSaving.value = true;
  try {
    const { data } = await updateMe({
      display_name: sessionStore.me?.display_name ?? sessionStore.currentUser.display_name,
      ...(sessionStore.me?.email ? { email: sessionStore.me.email } : {}),
      ...(isSystemPhonePlaceholder(sessionStore.me?.phone_country_code)
        ? {}
        : {
            phone_country_code: sessionStore.me?.phone_country_code ?? '',
            phone_number: sessionStore.me?.phone_number ?? '',
          }),
      primary_community_id: selectedBuildingID.value,
      primary_community_name: selectedBuildingName.value,
      residence_floor: selectedFloor.value,
      residence_unit: selectedUnitName.value,
      district_code: sessionStore.me?.district_code ?? '',
    });
    sessionStore.me = data.data;
    feedbackStore.pushToast('繳費單位已更新。', 'success');
  } catch (error) {
    feedbackStore.pushToast(readErrorMessage(error, '繳費單位儲存失敗，請稍後再試。'), 'error');
  } finally {
    isSaving.value = false;
  }
};

onMounted(async () => {
  try {
    if (!sessionStore.me) {
      await sessionStore.loadCurrentUser();
    }
    syncFromProfile();
    await loadBuildings();
    if (selectedBuildingID.value) {
      await loadUnits(selectedBuildingID.value);
      syncSelectedUnitFromProfile();
    }
  } finally {
    isSyncingProfileSelection = false;
  }
});

watch(selectedBuildingID, (nextValue, previousValue) => {
  if (nextValue === previousValue) {
    return;
  }
  if (!isSyncingProfileSelection) {
    selectedFloor.value = '';
    selectedUnitID.value = '';
  }
  void loadUnits(nextValue);
});

watch(selectedFloor, (nextValue, previousValue) => {
  if (nextValue !== previousValue && !isSyncingProfileSelection) {
    selectedUnitID.value = '';
  }
});
</script>

<template>
  <main class="pos-page">
    <section class="pos-page__header">
      <div>
        <p>Payments / Unit</p>
        <h1>單元</h1>
        <span>{{ currentLabel }}</span>
      </div>
      <BaseButton
        variant="secondary"
        size="md"
        :disabled="isLoading || isUnitsLoading"
        @click="loadBuildings"
      >
        {{ isLoading ? '載入中' : '刷新' }}
      </BaseButton>
    </section>

    <section class="pos-page__selector pos-page__selector--three">
      <label>
        <span>大廈</span>
        <AppGlassSelect
          v-model="selectedBuildingID"
          :options="buildingOptions"
          :disabled="isLoading || buildingOptions.length === 0"
        />
      </label>
      <label>
        <span>樓層</span>
        <AppGlassSelect
          v-model="selectedFloor"
          :options="floorOptions"
          :disabled="!selectedBuildingID || isUnitsLoading || floorOptions.length === 0"
        />
      </label>
      <label>
        <span>單位</span>
        <AppGlassSelect
          v-model="selectedUnitID"
          :options="unitOptions"
          :disabled="!selectedFloor || isUnitsLoading || unitOptions.length === 0"
        />
      </label>
    </section>

    <section class="pos-page__notice">
      <p>{{ currentLabel }}</p>
      <BaseButton
        variant="primary"
        size="md"
        :disabled="!canSave || isSaving"
        @click="handleSaveUnit"
      >
        {{ isSaving ? '儲存中' : '儲存單位' }}
      </BaseButton>
    </section>
  </main>
</template>

<style scoped>
@import '../payment-page.css';
</style>
