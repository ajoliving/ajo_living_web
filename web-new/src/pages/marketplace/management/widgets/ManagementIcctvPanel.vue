<!--
 * 管理中心 - ICCTV 面板。
 * 1. 對齊 docs 高保真參考的 admin-icctv 結構與樣式。
 * 2. 目前為 mock UI，大廈與鏡頭資料為靜態，後續再接後端監控 API。
 * 3. 支援大廈選擇、鏡頭列表、監控查看器與接入狀態側欄。
-->
<script setup lang="ts">
/*
 * ICCTV 面板邏輯。
 * 1. 大廈與鏡頭 mock 資料。
 * 2. 選擇大廈與鏡頭狀態。
 * 3. 接入狀態標籤。
 */
import { computed, ref } from 'vue';

// 1. 大廈 mock 資料
interface IcctvBuilding {
  code: string;
  name: string;
  cameraCount: number;
  onlineCount: number;
}

const buildings: IcctvBuilding[] = [
  { code: 'b1', name: '康睦庭園第二座', cameraCount: 8, onlineCount: 8 },
  { code: 'b2', name: '協和大廈', cameraCount: 6, onlineCount: 5 },
  { code: 'b3', name: '仁英大廈', cameraCount: 4, onlineCount: 4 },
];

// 2. 鏡頭 mock 資料
interface IcctvCamera {
  code: string;
  buildingCode: string;
  name: string;
  source: string;
  online: boolean;
}

const cameras: IcctvCamera[] = [
  { code: 'c1', buildingCode: 'b1', name: '大堂入口', source: 'Orange Pi · CAM-01', online: true },
  { code: 'c2', buildingCode: 'b1', name: '電梯大堂 1F', source: 'Orange Pi · CAM-02', online: true },
  { code: 'c3', buildingCode: 'b1', name: '後巷出口', source: 'Orange Pi · CAM-03', online: true },
  { code: 'c4', buildingCode: 'b2', name: '大堂入口', source: 'Orange Pi · CAM-01', online: true },
  { code: 'c5', buildingCode: 'b2', name: '停車場 B1', source: 'Orange Pi · CAM-02', online: false },
  { code: 'c6', buildingCode: 'b3', name: '大堂入口', source: 'Orange Pi · CAM-01', online: true },
];

// 3. 當前選擇
const selectedBuildingCode = ref<string>('b1');
const selectedCameraCode = ref<string>('');

// 4. 當前大廈
const selectedBuilding = computed<IcctvBuilding | undefined>(() =>
  buildings.find((item) => item.code === selectedBuildingCode.value),
);

// 5. 當前大廈鏡頭
const buildingCameras = computed<IcctvCamera[]>(() =>
  cameras.filter((item) => item.buildingCode === selectedBuildingCode.value),
);

// 6. 當前鏡頭
const selectedCamera = computed<IcctvCamera | undefined>(() =>
  cameras.find((item) => item.code === selectedCameraCode.value),
);

// 7. 接入狀態標籤
const accessTags = computed(() => [
  { label: 'Orange Pi', on: true },
  { label: 'iCCTV Service', on: true },
  { label: 'WebRTC', on: false },
  { label: 'RTSP', on: true },
]);

// 8. 選擇鏡頭
const selectCamera = (code: string): void => {
  selectedCameraCode.value = code;
};
</script>

<template>
  <div class="icctv-panel">
    <!-- 1. 標題列 -->
    <section class="staff-list-head">
      <div>
        <div class="staff-kicker">Staff</div>
        <h2 class="staff-title">ICCTV</h2>
        <p class="staff-desc">按大廈查看即時監控畫面，切換 Orange Pi 與鏡頭來源。</p>
      </div>
      <button
        type="button"
        class="work-action"
      >
        新窗口
      </button>
    </section>

    <!-- 2. 監控網格 -->
    <section class="icctv-manage-grid">
      <!-- 2.1 左側監控卡 -->
      <div class="work-card icctv-main-card">
        <!-- 2.1.1 大廈選擇 -->
        <div class="icctv-building-select">
          <select
            v-model="selectedBuildingCode"
            class="staff-select"
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
            @click="selectedCameraCode = ''"
          >
            重新載入
          </button>
        </div>

        <!-- 2.1.2 摘要 -->
        <div
          v-if="selectedBuilding"
          class="icctv-summary"
        >
          <div class="icctv-summary-card">
            <div class="icctv-summary-num">{{ selectedBuilding.cameraCount }}</div>
            <div class="icctv-summary-label">鏡頭總數</div>
          </div>
          <div class="icctv-summary-card">
            <div class="icctv-summary-num">{{ selectedBuilding.onlineCount }}</div>
            <div class="icctv-summary-label">在線鏡頭</div>
          </div>
          <div class="icctv-summary-card">
            <div class="icctv-summary-num">{{ selectedBuilding.cameraCount - selectedBuilding.onlineCount }}</div>
            <div class="icctv-summary-label">離線鏡頭</div>
          </div>
        </div>

        <!-- 2.1.3 大廈資訊 -->
        <div
          v-if="selectedBuilding"
          class="icctv-building-meta"
        >
          <strong>{{ selectedBuilding.name }}</strong>
          <span>共 {{ selectedBuilding.cameraCount }} 個鏡頭，{{ selectedBuilding.onlineCount }} 個在線。</span>
        </div>

        <!-- 2.1.4 鏡頭列表 -->
        <div class="icctv-camera-list">
          <button
            v-for="item in buildingCameras"
            :key="item.code"
            type="button"
            class="icctv-camera-btn"
            :class="{ on: selectedCameraCode === item.code }"
            @click="selectCamera(item.code)"
          >
            <span>
              {{ item.name }}
              <small>{{ item.source }}</small>
            </span>
            <span
              class="icctv-camera-status"
              :class="item.online ? 'on' : 'off'"
            >
              {{ item.online ? '在線' : '離線' }}
            </span>
          </button>
          <div
            v-if="buildingCameras.length === 0"
            class="icctv-camera-empty"
          >
            目前大廈未有鏡頭資料。
          </div>
        </div>

        <!-- 2.1.5 查看器工具列 -->
        <div class="icctv-viewer-bar">
          <div class="icctv-viewer-meta">
            {{ selectedCamera ? `${selectedCamera.name} · ${selectedCamera.source}` : '尚未選擇鏡頭' }}
          </div>
          <div class="icctv-viewer-actions">
            <button
              type="button"
              class="work-mini-btn"
            >
              新窗口
            </button>
          </div>
        </div>

        <!-- 2.1.6 監控查看器 -->
        <div class="icctv-viewer">
          <iframe
            v-if="selectedCamera"
            title="ICCTV 即時監控"
            allow="autoplay; fullscreen; encrypted-media; picture-in-picture"
          />
          <div
            v-else
            class="icctv-empty"
          >
            <strong>未選擇鏡頭</strong>
            <span>選擇大廈與鏡頭後，監控畫面會顯示在此處。</span>
          </div>
        </div>
      </div>

      <!-- 2.2 右側側欄 -->
      <aside class="icctv-manage-side">
        <div class="icctv-manage-panel">
          <h4>接入狀態</h4>
          <div class="icctv-tag-row">
            <span
              v-for="tag in accessTags"
              :key="tag.label"
              class="icctv-tag"
              :class="{ on: tag.on }"
            >
              {{ tag.label }}
            </span>
          </div>
          <p class="icctv-note">Staff 先選擇大廈，再選擇鏡頭。畫面使用後台返回的監控 URL 嵌入。</p>
        </div>
        <div class="icctv-manage-panel">
          <h4>嵌入方式</h4>
          <p class="icctv-note">目前使用 iframe 預覽。若來源頁限制嵌入或瀏覽器阻擋混合內容，可使用新窗口開啟，後續再接專用 WebRTC 播放器。</p>
        </div>
      </aside>
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

/* 4. 通用卡片 */
.work-card {
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: #fff;
  padding: 16px;
}

/* 5. 網格佈局 */
.icctv-manage-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 320px;
  gap: 12px;
  align-items: start;
}

/* 6. 大廈選擇 */
.icctv-building-select {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
}

.staff-select {
  flex: 1;
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

/* 7. 摘要 */
.icctv-summary {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
  margin-top: 14px;
}

.icctv-summary-card {
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: var(--sur);
  padding: 14px;
}

.icctv-summary-num {
  color: var(--brand);
  font-size: 26px;
  font-weight: 800;
  line-height: 1;
}

.icctv-summary-label {
  margin-top: 6px;
  color: var(--ink-3);
  font-size: 12px;
  font-weight: 700;
}

/* 8. 大廈資訊 */
.icctv-building-meta {
  margin-top: 12px;
  border-top: 1px solid var(--bdr);
  padding-top: 12px;
}

.icctv-building-meta strong {
  display: block;
  color: var(--ink);
  font-size: 14px;
  font-weight: 800;
}

.icctv-building-meta span {
  display: block;
  margin-top: 4px;
  color: var(--ink-3);
  font-size: 12px;
}

/* 9. 鏡頭列表 */
.icctv-camera-list {
  display: grid;
  gap: 8px;
  margin-top: 12px;
}

.icctv-camera-btn {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: var(--sur);
  color: var(--ink);
  cursor: pointer;
  font-family: inherit;
  font-size: 13px;
  font-weight: 800;
  padding: 10px 12px;
  text-align: left;
}

.icctv-camera-btn.on {
  border-color: var(--brand);
  background: var(--brand-light);
  color: var(--brand);
}

.icctv-camera-btn small {
  display: block;
  margin-top: 3px;
  color: var(--ink-3);
  font-size: 11px;
  font-weight: 700;
}

.icctv-camera-status {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 800;
  padding: 4px 8px;
  white-space: nowrap;
}

.icctv-camera-status.on {
  background: var(--success-bg);
  color: var(--success);
}

.icctv-camera-status.off {
  background: var(--sur-2);
  color: var(--ink-4);
}

.icctv-camera-empty {
  color: var(--ink-3);
  font-size: 12px;
  font-weight: 600;
  text-align: center;
  padding: 18px 12px;
}

/* 10. 查看器工具列 */
.icctv-viewer-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  padding: 12px 0 0;
}

.icctv-viewer-meta {
  color: var(--ink-3);
  font-size: 12px;
  font-weight: 700;
}

.icctv-viewer-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

/* 11. 監控查看器 */
.icctv-viewer {
  position: relative;
  min-height: 520px;
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: #111827;
  overflow: hidden;
  margin-top: 10px;
}

.icctv-viewer iframe {
  display: block;
  width: 100%;
  height: 520px;
  border: 0;
  background: #111827;
}

.icctv-empty {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;
  gap: 8px;
  background: linear-gradient(160deg, rgba(17, 24, 39, 0.88), rgba(17, 24, 39, 0.65));
  color: #E5E7EB;
  text-align: center;
  padding: 20px;
}

.icctv-empty strong {
  font-size: 18px;
  font-weight: 800;
}

.icctv-empty span {
  font-size: 13px;
  color: #94A3B8;
  font-weight: 600;
  max-width: 280px;
  line-height: 1.6;
}

/* 12. 右側側欄 */
.icctv-manage-side {
  display: grid;
  gap: 12px;
}

.icctv-manage-panel {
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: var(--sur);
  padding: 14px;
}

.icctv-manage-panel h4 {
  margin: 0 0 10px;
  color: var(--ink);
  font-size: 14px;
  font-weight: 800;
}

.icctv-tag-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.icctv-tag {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  background: var(--sur-2);
  color: var(--ink-2);
  font-size: 11px;
  font-weight: 800;
  padding: 5px 9px;
}

.icctv-tag.on {
  background: var(--brand-light);
  color: var(--brand);
}

.icctv-note {
  margin: 12px 0 0;
  color: var(--ink-3);
  font-size: 12px;
  line-height: 1.6;
}

/* 13. 響應式 */
@media (max-width: 980px) {
  .icctv-manage-grid {
    grid-template-columns: 1fr;
  }

  .icctv-summary {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 560px) {
  .staff-list-head {
    flex-direction: column;
    align-items: flex-start;
  }

  .icctv-viewer {
    min-height: 320px;
  }

  .icctv-viewer iframe {
    height: 320px;
  }
}
</style>
