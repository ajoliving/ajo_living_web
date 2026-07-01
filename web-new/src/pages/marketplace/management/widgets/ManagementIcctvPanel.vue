<!--
 * 管理中心 - ICCTV 面板。
 * 1. 對齊 docs 高保真參考的 admin-icctv 結構與樣式。
 * 2. 讀取目前 Staff 可見大廈的真實 iCCTV 鏡頭資料。
 * 3. 支援大廈選擇、鏡頭列表、表格展開與多鏡頭同時查看。
-->
<script setup lang="ts">
/*
 * ICCTV 面板邏輯。
 * 1. 讀取可見大廈與 iCCTV 鏡頭資料。
 * 2. 切換大廈與展開鏡頭狀態。
 * 3. 開啟單個鏡頭新窗口。
 */
import { computed, onMounted, ref } from 'vue';

import {
  fetchMemberICCTVPublicCameras,
  fetchMemberPosBuildings,
  type ICCTVCameraSummary,
  type ICCTVPublicCameraResponse,
} from '@/httpapis/building';
import type { PosBuilding } from '@/model/community';

interface IcctvBuildingOption {
  code: string;
  name: string;
}

const loading = ref(false);
const errorMessage = ref('');
const selectedBuildingCode = ref('');
const expandedCameraIDs = ref<string[]>([]);
const icctvProfile = ref<ICCTVPublicCameraResponse | null>(null);
const posBuildings = ref<PosBuilding[]>([]);

// 1. 取得大廈 ID
const posBuildingID = (building: PosBuilding): string =>
  String(building.building_id || building.id || '').trim();

// 2. 取得大廈顯示名稱
const posBuildingName = (building: PosBuilding): string =>
  String(building.buildname_chi || building.buildname || building.name || posBuildingID(building)).trim();

// 3. 建立大廈名稱索引
const buildingNameMap = computed<Record<string, string>>(() =>
  posBuildings.value.reduce<Record<string, string>>((result, building) => {
    const buildingID = posBuildingID(building);
    if (buildingID) {
      result[buildingID] = posBuildingName(building);
    }
    return result;
  }, {}),
);

// 4. 建立可選大廈
const buildings = computed<IcctvBuildingOption[]>(() =>
  (icctvProfile.value?.building_options ?? []).map((buildingID) => ({
    code: buildingID,
    name: buildingNameMap.value[buildingID] || buildingID,
  })),
);

// 5. 當前大廈
const selectedBuilding = computed<IcctvBuildingOption | undefined>(() =>
  buildings.value.find((item) => item.code === selectedBuildingCode.value),
);

// 6. 當前大廈鏡頭
const buildingCameras = computed<ICCTVCameraSummary[]>(() => icctvProfile.value?.cameras ?? []);
const totalCameraCount = computed(() => buildingCameras.value.length);
const onlineCameraCount = computed(() => buildingCameras.value.filter((item) => item.is_active && item.url).length);
const offlineCameraCount = computed(() => Math.max(0, totalCameraCount.value - onlineCameraCount.value));

// 7. 取得鏡頭名稱
const cameraName = (camera: ICCTVCameraSummary, index: number): string => {
  const match = String(camera.channel ?? '').match(/^channel(\d+)$/i);
  return `鏡頭 ${match?.[1] ?? index + 1}`;
};

// 8. 取得鏡頭狀態
const cameraStatusText = (camera: ICCTVCameraSummary): string =>
  camera.is_active && camera.url ? '可查看' : '不可查看';

// 9. 取得鏡頭 iframe 標題
const cameraFrameTitle = (camera: ICCTVCameraSummary, index: number): string =>
  `${cameraName(camera, index)} 即時監控`;

// 10. 判斷鏡頭是否已展開
const isCameraExpanded = (cameraID: string): boolean => expandedCameraIDs.value.includes(cameraID);

// 11. 展開或收起鏡頭
const toggleCamera = (cameraID: string): void => {
  expandedCameraIDs.value = isCameraExpanded(cameraID)
    ? expandedCameraIDs.value.filter((id) => id !== cameraID)
    : [...expandedCameraIDs.value, cameraID];
};

// 12. 開啟鏡頭新窗口
const openCameraWindow = (camera: ICCTVCameraSummary): void => {
  if (!camera?.url) return;
  window.open(camera.url, '_blank', 'noopener');
};

// 13. 讀取大廈名稱
const loadBuildingNames = async (): Promise<void> => {
  try {
    posBuildings.value = await fetchMemberPosBuildings();
  } catch {
    posBuildings.value = [];
  }
};

// 14. 讀取 iCCTV 鏡頭
const loadICCTV = async (buildingID = selectedBuildingCode.value): Promise<void> => {
  loading.value = true;
  errorMessage.value = '';
  try {
    const result = await fetchMemberICCTVPublicCameras(buildingID || undefined);
    icctvProfile.value = result;
    selectedBuildingCode.value = result.selected_building_id || buildingID || '';
    expandedCameraIDs.value = [];
  } catch (error) {
    console.error(error);
    errorMessage.value = '視像監控資料載入失敗';
    icctvProfile.value = null;
    expandedCameraIDs.value = [];
  } finally {
    loading.value = false;
  }
};

// 15. 切換大廈
const changeBuilding = (): void => {
  void loadICCTV(selectedBuildingCode.value);
};

onMounted(() => {
  void loadBuildingNames();
  void loadICCTV();
});
</script>

<template>
  <div class="icctv-panel">
    <!-- 1. 標題列 -->
    <section class="staff-list-head">
      <div>
        <div class="staff-kicker">Staff</div>
        <h2 class="staff-title">ICCTV</h2>
        <p class="staff-desc">按大廈查看即時監控畫面。</p>
      </div>
      <button
        type="button"
        class="work-action"
        :disabled="loading"
        @click="loadICCTV()"
      >
        {{ loading ? '載入中' : '重新整理' }}
      </button>
    </section>

    <!-- 2. 監控表格 -->
    <section class="work-card icctv-table-card">
      <div class="icctv-profile-head">
        <div>
          <h3>{{ selectedBuilding?.name || '未選擇大廈' }}</h3>
          <p>共 {{ totalCameraCount }} 個鏡頭，{{ onlineCameraCount }} 個在線，{{ offlineCameraCount }} 個離線。</p>
        </div>
        <div class="icctv-building-select">
          <select
            v-model="selectedBuildingCode"
            class="staff-select"
            :disabled="loading || buildings.length === 0"
            @change="changeBuilding"
          >
            <option
              v-for="item in buildings"
              :key="item.code"
              :value="item.code"
            >
              {{ item.name }}
            </option>
          </select>
          <button
            type="button"
            class="work-mini-btn"
            :disabled="loading"
            @click="loadICCTV()"
          >
            {{ loading ? '載入中' : '重新載入' }}
          </button>
        </div>
      </div>

      <div
        v-if="errorMessage"
        class="icctv-table-message icctv-table-error"
      >
        {{ errorMessage }}
      </div>

      <div class="icctv-table-wrap">
        <table class="work-table icctv-table">
          <thead>
            <tr>
              <th>鏡頭</th>
              <th>狀態</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <template
              v-for="(camera, index) in buildingCameras"
              :key="camera.id"
            >
              <tr :class="{ expanded: isCameraExpanded(camera.id) }">
                <td>
                  <div class="icctv-camera-title">{{ cameraName(camera, index) }}</div>
                  <div class="icctv-camera-channel">{{ camera.channel }}</div>
                </td>
                <td>
                  <span
                    class="icctv-table-status"
                    :class="camera.is_active && camera.url ? 'on' : 'off'"
                  >
                    {{ cameraStatusText(camera) }}
                  </span>
                </td>
                <td>
                  <div class="icctv-row-actions">
                    <button
                      type="button"
                      class="work-mini-btn primary"
                      :disabled="!camera.url"
                      @click="toggleCamera(camera.id)"
                    >
                      {{ isCameraExpanded(camera.id) ? '收起' : '查看' }}
                    </button>
                    <button
                      type="button"
                      class="work-mini-btn"
                      :disabled="!camera.url"
                      @click="openCameraWindow(camera)"
                    >
                      新窗口
                    </button>
                  </div>
                </td>
              </tr>
              <tr
                v-if="isCameraExpanded(camera.id)"
                class="icctv-expanded-row"
              >
                <td colspan="3">
                  <div class="icctv-inline-viewer">
                    <iframe
                      :key="`${camera.id}-frame`"
                      :src="camera.url"
                      :title="cameraFrameTitle(camera, index)"
                      allow="autoplay; fullscreen; encrypted-media; picture-in-picture"
                    />
                  </div>
                </td>
              </tr>
            </template>
            <tr v-if="buildingCameras.length === 0">
              <td
                class="icctv-table-message"
                colspan="3"
              >
                {{ loading ? '正在載入鏡頭資料。' : errorMessage || '目前大廈未有鏡頭資料。' }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </div>
</template>

<style scoped>
/* 1. 面板容器 */
.icctv-panel {
  display: grid;
  gap: 14px;
}

/* 2. 標題列 */
.staff-list-head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 12px;
}

.staff-kicker {
  color: var(--ink-3);
  font-size: 10px;
  font-weight: 800;
  letter-spacing: 1.6px;
  text-transform: uppercase;
}

.staff-title {
  margin: 6px 0 0;
  color: var(--ink);
  font-family: var(--font-serif);
  font-size: 30px;
  font-weight: 400;
  line-height: 1.1;
}

.staff-desc {
  margin: 6px 0 0;
  color: var(--ink-3);
  font-size: 13px;
  font-weight: 600;
  line-height: 1.5;
}

/* 3. 通用按鈕 */
.work-action {
  border: 0;
  border-radius: 6px;
  background: var(--accent);
  color: #fff;
  cursor: pointer;
  font-family: inherit;
  font-size: 12px;
  font-weight: 600;
  padding: 10px 14px;
  white-space: nowrap;
}

.work-mini-btn {
  border: 1px solid var(--bdr);
  border-radius: 6px;
  background: var(--sur);
  color: var(--ink);
  cursor: pointer;
  font-family: inherit;
  font-size: 12px;
  font-weight: 600;
  padding: 7px 10px;
}

.work-action:disabled,
.work-mini-btn:disabled,
.staff-select:disabled {
  cursor: not-allowed;
  opacity: 0.52;
}

/* 4. 通用卡片 */
.work-card {
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: #fff;
  padding: 16px;
}

/* 5. 表格卡片 */
.icctv-table-card {
  padding: 0;
  overflow: hidden;
}

.icctv-profile-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 18px;
  border-bottom: 1px solid var(--bdr);
  background: #fff;
  padding: 18px;
}

.icctv-profile-head h3 {
  margin: 0;
  color: var(--ink);
  font-size: 18px;
  font-weight: 900;
  line-height: 1.35;
}

.icctv-profile-head p {
  margin: 6px 0 0;
  color: var(--ink-3);
  font-size: 13px;
  font-weight: 700;
  line-height: 1.5;
}

/* 6. 大廈選擇 */
.icctv-building-select {
  display: flex;
  align-items: center;
  gap: 10px;
  width: min(420px, 100%);
}

.staff-select {
  flex: 1;
  min-width: 0;
  height: 36px;
  border: 1px solid var(--bdr);
  border-radius: 6px;
  background: var(--sur);
  color: var(--ink);
  font-family: inherit;
  font-size: 13px;
  font-weight: 700;
  padding: 0 12px;
}

.icctv-building-select .work-mini-btn {
  white-space: nowrap;
}

/* 7. 鏡頭表格 */
.work-table {
  width: 100%;
  border-collapse: collapse;
}

.work-table th,
.work-table td {
  border-bottom: 1px solid var(--bdr);
  padding: 14px 16px;
  text-align: left;
  vertical-align: middle;
}

.work-table th {
  background: var(--sur);
  color: var(--ink-3);
  font-size: 11px;
  font-weight: 900;
  text-transform: uppercase;
}

.icctv-table-wrap {
  overflow-x: auto;
}

.icctv-table {
  min-width: 760px;
  table-layout: fixed;
}

.icctv-table th:nth-child(1),
.icctv-table td:nth-child(1) {
  width: 44%;
}

.icctv-table th:nth-child(2),
.icctv-table td:nth-child(2) {
  width: 20%;
}

.icctv-table th:nth-child(3),
.icctv-table td:nth-child(3) {
  width: 36%;
}

.icctv-table tbody tr.expanded > td {
  background: var(--brand-light);
}

.icctv-camera-title {
  color: var(--ink);
  font-size: 13px;
  font-weight: 900;
}

.icctv-camera-channel {
  margin-top: 4px;
  color: var(--ink-3);
  font-size: 12px;
  font-weight: 700;
}

.icctv-table-status {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 900;
  padding: 5px 9px;
  white-space: nowrap;
}

.icctv-table-status.on {
  background: var(--success-bg);
  color: var(--success);
}

.icctv-table-status.off {
  background: var(--sur-2);
  color: var(--ink-4);
}

.icctv-row-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.work-mini-btn.primary {
  border-color: var(--brand);
  background: var(--brand-light);
  color: var(--brand);
}

.icctv-expanded-row td {
  background: #fff;
  padding: 0;
}

.icctv-inline-viewer {
  border-top: 1px solid var(--bdr);
  background: #0F172A;
}

.icctv-inline-viewer iframe {
  display: block;
  width: 100%;
  height: 520px;
  border: 0;
  background: #0F172A;
}

.icctv-table-message {
  color: var(--ink-3);
  font-size: 12px;
  font-weight: 600;
  text-align: center;
  padding: 18px 12px;
}

.icctv-table-error {
  border: 1px solid rgba(186, 26, 26, 0.2);
  border-radius: 0;
  border-width: 0 0 1px;
  background: rgba(186, 26, 26, 0.06);
  color: #ba1a1a;
}

/* 8. 響應式 */
@media (max-width: 980px) {
  .icctv-profile-head {
    align-items: stretch;
    flex-direction: column;
  }

  .icctv-building-select {
    width: 100%;
  }
}

@media (max-width: 560px) {
  .staff-list-head {
    flex-direction: column;
    align-items: flex-start;
  }

  .icctv-building-select {
    align-items: stretch;
    flex-direction: column;
  }

  .icctv-inline-viewer iframe {
    height: 320px;
  }
}
</style>
