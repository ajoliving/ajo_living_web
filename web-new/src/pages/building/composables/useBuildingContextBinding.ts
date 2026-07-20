/*
 * 我的大廈 - 目前物業切換狀態。
 * 1. 將 iSmart 已授權單位整理成「大廈 / 樓層 / 單位」單一選項。
 * 2. 保存會員目前使用的完整物業上下文。
 * 3. 向桌面及手機原生下拉入口提供同一份狀態。
 */
import { computed, ref, type Ref } from 'vue';
import { useI18n } from 'vue-i18n';

import { fetchMemberPosBuildingUnits } from '@/httpapis/building';
import { updateMe } from '@/httpapis/me';
import type { PosBuilding, PosBuildingUnit } from '@/model/community';
import { useSessionStore } from '@/stores/session';

import {
  buildBuildingNameMap,
  indexedBuildingName,
  memberCommunityName,
  preferredMemberBuildingID,
} from './building-display';

interface BuildingContextOption {
  buildingID: string;
  floor: string;
  label: string;
  unit: string;
  value: string;
}

const unassignedFloorValue = '__unassigned__';

// 1. 正規化 POS 識別碼與權限列表
const digitsOnly = (value: unknown): string => String(value ?? '').replace(/\D/g, '');
const normalizeUnitIDs = (values: string[] | undefined): string[] => Array.from(new Set(
  (values ?? [])
    .flatMap((item) => String(item ?? '').split(/[,，\n\r]+/))
    .map((item) => digitsOnly(item).slice(0, 11))
    .filter((item) => item.length === 11),
));

// 2. 讀取 POS 單位欄位
const unitID = (item: PosBuildingUnit): string => digitsOnly(item.unit_id ?? item.id).slice(0, 11);
const unitFloor = (item: PosBuildingUnit): string => String(item.floor ?? '').trim() || unassignedFloorValue;
const unitName = (item: PosBuildingUnit): string => String(item.unit ?? item.unit_name ?? item.name ?? '').trim();
const compareCodes = (left: string, right: string): number => left.localeCompare(right, 'en', {
  numeric: true,
  sensitivity: 'base',
});

// 3. useBuildingContextBinding manages one complete authorized-property selector.
export const useBuildingContextBinding = (buildingDirectory: Ref<PosBuilding[]>) => {
  const sessionStore = useSessionStore();
  const { t, locale } = useI18n();
  const selectedUnitID = ref('');
  const units = ref<PosBuildingUnit[]>([]);
  const loading = ref(false);
  const saving = ref(false);
  const status = ref<'idle' | 'success' | 'error'>('idle');
  const errorKey = ref('');

  const allowedUnitIDs = computed(() => {
    const ismartUnits = normalizeUnitIDs(sessionStore.me?.ismart_msg?.client_building_flat_units_permissions);
    return ismartUnits.length > 0 ? ismartUnits : normalizeUnitIDs(sessionStore.me?.bound_flat_unit_ids);
  });
  const allowedBuildingIDs = computed(() => Array.from(new Set(
    allowedUnitIDs.value.map((item) => item.slice(0, 7)),
  )));
  const buildingNames = computed(() => buildBuildingNameMap(
    buildingDirectory.value,
    sessionStore.me?.primary_community,
    locale.value,
  ));

  const propertyOptions = computed<BuildingContextOption[]>(() => {
    const permissions = new Set(allowedUnitIDs.value);
    return units.value
      .map((item) => {
        const value = unitID(item);
        const buildingID = value.slice(0, 7);
        const floor = unitFloor(item);
        const unit = unitName(item);
        const buildingName = indexedBuildingName(buildingNames.value, buildingID)
          || (buildingID === sessionStore.me?.primary_community?.public_id
            ? memberCommunityName(sessionStore.me?.primary_community, locale.value)
            : '')
          || t('building.common.buildingCodeLabel', { id: buildingID });
        return {
          value,
          buildingID,
          floor,
          unit,
          label: [
            buildingName,
            floor === unassignedFloorValue ? t('building.context.unassignedFloor') : floor,
            unit,
          ].filter(Boolean).join(' / '),
        };
      })
      .filter((item) => permissions.has(item.value) && item.unit)
      .sort((left, right) => compareCodes(left.label, right.label));
  });

  const selectedProperty = computed(() => propertyOptions.value.find((item) => item.value === selectedUnitID.value));
  const canSave = computed(() => Boolean(selectedProperty.value));
  const currentLabel = computed(() => {
    const savedUnitIDs = normalizeUnitIDs(sessionStore.me?.bound_flat_unit_ids);
    const savedOption = propertyOptions.value.find((item) => savedUnitIDs.includes(item.value));
    if (savedOption) return savedOption.label;

    const buildingID = preferredMemberBuildingID(sessionStore.me);
    const buildingName = indexedBuildingName(buildingNames.value, buildingID)
      || memberCommunityName(sessionStore.me?.primary_community, locale.value)
      || buildingID;
    return [buildingName, sessionStore.me?.residence_floor, sessionStore.me?.residence_unit]
      .map((item) => String(item ?? '').trim())
      .filter(Boolean)
      .join(' / ');
  });

  // 4. 載入所有已授權大廈的單位並只保留授權項目
  const initialize = async (): Promise<void> => {
    loading.value = true;
    status.value = 'idle';
    errorKey.value = '';
    try {
      const results = await Promise.allSettled(
        allowedBuildingIDs.value.map((buildingID) => fetchMemberPosBuildingUnits(buildingID)),
      );
      const unitMap = new Map<string, PosBuildingUnit>();
      results.forEach((result) => {
        if (result.status !== 'fulfilled') return;
        result.value.forEach((item) => {
          const value = unitID(item);
          if (allowedUnitIDs.value.includes(value)) unitMap.set(value, item);
        });
      });
      units.value = Array.from(unitMap.values());
      if (allowedUnitIDs.value.length > 0 && units.value.length === 0) {
        errorKey.value = 'building.context.unitLoadError';
      }

      const savedUnitIDs = normalizeUnitIDs(sessionStore.me?.bound_flat_unit_ids);
      const savedOption = propertyOptions.value.find((item) => savedUnitIDs.includes(item.value))
        || propertyOptions.value.find((item) => (
          item.buildingID === preferredMemberBuildingID(sessionStore.me)
          && (item.floor === unassignedFloorValue ? '' : item.floor) === (sessionStore.me?.residence_floor?.trim() ?? '')
          && item.unit === (sessionStore.me?.residence_unit?.trim() ?? '')
        ));
      selectedUnitID.value = savedOption?.value ?? propertyOptions.value[0]?.value ?? '';
    } finally {
      loading.value = false;
    }
  };

  // 5. 保存完整物業選項至會員資料
  const save = async (): Promise<boolean> => {
    if (!selectedProperty.value) {
      status.value = 'error';
      errorKey.value = 'building.context.selectionRequired';
      return false;
    }

    saving.value = true;
    status.value = 'idle';
    errorKey.value = '';
    try {
      const option = selectedProperty.value;
      const buildingName = indexedBuildingName(buildingNames.value, option.buildingID)
        || memberCommunityName(sessionStore.me?.primary_community, locale.value)
        || option.buildingID;
      const displayName = sessionStore.me?.display_name?.trim() ?? '';
      const { data } = await updateMe({
        ...(displayName ? { display_name: displayName } : {}),
        primary_community_id: option.buildingID,
        primary_community_name: buildingName,
        bound_building_ids: [option.buildingID],
        bound_flat_unit_ids: [option.value],
        residence_floor: option.floor === unassignedFloorValue ? '' : option.floor,
        residence_unit: option.unit,
      });
      sessionStore.me = data.data;
      status.value = 'success';
      return true;
    } catch {
      status.value = 'error';
      errorKey.value = 'building.context.saveError';
      return false;
    } finally {
      saving.value = false;
    }
  };

  return {
    selectedUnitID,
    propertyOptions,
    loading,
    saving,
    status,
    errorKey,
    canSave,
    currentLabel,
    selectedProperty,
    initialize,
    save,
  };
};
