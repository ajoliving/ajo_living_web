<!--
 * 管理中心頁。
 * 1. 左側子導航：家具列表、樓盤列表、住宅列表、ICCTV、會員列表、廣告列表、廣告設定、積分流水。
 * 2. 右側依當前分頁渲染對應的管理介面（表格、篩選、分頁、ICCTV 查看器、廣告設定）。
 * 3. 全部使用靜態 mock 資料，無 API 呼叫。
-->
<script setup lang="ts">
import { computed, ref } from 'vue';
import { useRouter } from 'vue-router';

import AppIcon from '@/shared/components/base/AppIcon.vue';

type ManagementTab =
  | 'admin-furniture'
  | 'admin-properties'
  | 'admin-homes'
  | 'admin-icctv'
  | 'admin-members'
  | 'admin-ads'
  | 'admin-ad-settings'
  | 'admin-points';

type ManagementIconName = 'layout-list' | 'home' | 'browse' | 'shield' | 'user' | 'plus-square' | 'palette' | 'wallet';

interface ManagementNavItem {
  key: ManagementTab;
  label: string;
  icon: ManagementIconName;
}

interface FurnitureRow {
  title: string;
  code: string;
  publisher: string;
  status: string;
  statusGood: boolean;
  publishedAt: string;
  expireAt: string;
  updatedAt: string;
  actions: string[];
}

interface PropertyRow {
  title: string;
  code: string;
  publisher: string;
  status: string;
  statusGood: boolean;
  publishedAt: string;
  expireAt: string;
  updatedAt: string;
  actions: string[];
}

interface HomeRow {
  name: string;
  sub: string;
  unit: string;
  resident: string;
  status: string;
  statusGood: boolean;
  updatedAt: string;
  actions: string[];
}

interface MemberRow {
  name: string;
  sub: string;
  type: string;
  role: string;
  status: string;
  statusGood: boolean;
  source: string;
  actions: string[];
}

interface AdRow {
  title: string;
  code: string;
  publisher: string;
  status: string;
  statusGood: boolean;
  publishedAt: string;
  expireAt: string;
  updatedAt: string;
  actions: string[];
}

interface AdSourceItem {
  thumb: string;
  title: string;
  meta: string;
}

interface AdSlotItem {
  thumb: string;
  name: string;
  status: string;
  desc: string;
}

interface PointRow {
  member: string;
  type: string;
  source: string;
  points: string;
  date: string;
}

interface IcctvBuilding {
  id: string;
  name: string;
}

interface IcctvCamera {
  id: string;
  name: string;
  location: string;
}

const router = useRouter();

// 1. 建立左側導航項目
const navItems: ManagementNavItem[] = [
  { key: 'admin-furniture', label: '家具列表', icon: 'layout-list' },
  { key: 'admin-properties', label: '樓盤列表', icon: 'home' },
  { key: 'admin-homes', label: '住宅列表', icon: 'browse' },
  { key: 'admin-icctv', label: 'ICCTV', icon: 'shield' },
  { key: 'admin-members', label: '會員列表', icon: 'user' },
  { key: 'admin-ads', label: '廣告列表', icon: 'layout-list' },
  { key: 'admin-ad-settings', label: '廣告設定', icon: 'palette' },
  { key: 'admin-points', label: '積分流水', icon: 'wallet' },
];

// 2. 當前選中的分頁
const activeTab = ref<ManagementTab>('admin-furniture');

// 3. 切換分頁
const switchTab = (key: ManagementTab): void => {
  activeTab.value = key;
};

// 4. 家具列表 mock 資料
const furnitureRows: FurnitureRow[] = [
  {
    title: '北歐實木餐桌',
    code: '01KRDJYNAMBP...',
    publisher: 'Admin',
    status: '上架中',
    statusGood: true,
    publishedAt: '2026年5月12日',
    expireAt: '2026年9月2日',
    updatedAt: '2026年6月4日',
    actions: ['下架', '續期'],
  },
  {
    title: 'LG 洗衣機 8kg',
    code: '01KOC7Z9STR5...',
    publisher: 'Admin',
    status: '未售出',
    statusGood: false,
    publishedAt: '2026年4月29日',
    expireAt: '2026年9月2日',
    updatedAt: '2026年6月4日',
    actions: ['下架', '續期'],
  },
  {
    title: '三色短毛貓領養',
    code: '01KSHP8BNHBT...',
    publisher: '抽C',
    status: '已過期',
    statusGood: false,
    publishedAt: '2026年5月26日',
    expireAt: '2026年6月9日',
    updatedAt: '2026年6月9日',
    actions: ['上架', '續期'],
  },
];

// 5. 樓盤列表 mock 資料
const propertyRows: PropertyRow[] = [
  {
    title: '佐敦高級住宅',
    code: '01KPROPERTY01...',
    publisher: 'Admin',
    status: '上架中',
    statusGood: true,
    publishedAt: '2026年5月22日',
    expireAt: '2026年9月2日',
    updatedAt: '2026年6月4日',
    actions: ['下架', '續期'],
  },
  {
    title: '沙田第一城 3房',
    code: '01KPROPERTY02...',
    publisher: 'test',
    status: '草稿',
    statusGood: false,
    publishedAt: '-',
    expireAt: '-',
    updatedAt: '2026年5月20日',
    actions: ['上架', '續期'],
  },
];

// 6. 住宅列表 mock 資料
const homeRows: HomeRow[] = [
  {
    name: '康睦庭園第二座',
    sub: '服務式住宅',
    unit: '02 / D',
    resident: 'patrick',
    status: '已綁定',
    statusGood: true,
    updatedAt: '2026年6月5日',
    actions: ['查看', '編輯'],
  },
  {
    name: 'Harbour Residence',
    sub: '住戶申請',
    unit: '08 / A',
    resident: '待確認',
    status: '待確認',
    statusGood: false,
    updatedAt: '2026年5月22日',
    actions: ['確認', '查看'],
  },
];

// 7. 會員列表 mock 資料
const memberRows: MemberRow[] = [
  {
    name: 'patrick',
    sub: 'patrick@example.com',
    type: 'user',
    role: 'staff',
    status: 'active',
    statusGood: true,
    source: 'isMart 同步',
    actions: ['查看', '停用'],
  },
  {
    name: 'chan@example.com',
    sub: 'owner',
    type: 'owner',
    role: 'resident',
    status: 'active',
    statusGood: true,
    source: 'AJO Living',
    actions: ['查看', '編輯'],
  },
];

// 8. 廣告列表 mock 資料
const adRows: AdRow[] = [
  {
    title: '社區家具回收',
    code: 'AD-FURNITURE-01',
    publisher: 'Admin',
    status: '投放中',
    statusGood: true,
    publishedAt: '2026年6月1日',
    expireAt: '2026年7月1日',
    updatedAt: '今天 11:00',
    actions: ['下架', '續期'],
  },
  {
    title: '住宅按揭諮詢',
    code: 'AD-PROPERTY-02',
    publisher: 'Admin',
    status: '草稿',
    statusGood: false,
    publishedAt: '-',
    expireAt: '-',
    updatedAt: '昨天 09:45',
    actions: ['上架', '續期'],
  },
];

// 9. 廣告素材 mock 資料
const adSources: AdSourceItem[] = [
  { thumb: '1', title: '社區家具回收', meta: '家具側欄 · 投放中' },
  { thumb: '2', title: '住宅按揭諮詢', meta: '樓盤側欄 · 草稿' },
  { thumb: '3', title: '搬屋預約', meta: '樓盤側欄 · 排期中' },
];

// 10. 廣告位 mock 資料
const adSlots: AdSlotItem[] = [
  { thumb: '1', name: '家具位置 1', status: '未設定', desc: '家具市集右側 16:9 廣告位' },
  { thumb: '2', name: '家具位置 2', status: '未設定', desc: '家具市集右側 16:9 廣告位' },
  { thumb: '3', name: '家具位置 3', status: '未設定', desc: '家具市集右側 16:9 廣告位' },
  { thumb: '4', name: '家具位置 4', status: '未設定', desc: '家具市集右側 9:16 廣告位' },
  { thumb: '5', name: '家具位置 5', status: '未設定', desc: '家具市集右側 9:16 廣告位' },
  { thumb: '6', name: '樓盤位置 1', status: '未設定', desc: '樓盤列表右側 16:9 廣告位' },
  { thumb: '7', name: '樓盤位置 2', status: '未設定', desc: '樓盤列表右側 16:9 廣告位' },
  { thumb: '8', name: '樓盤位置 3', status: '未設定', desc: '樓盤列表右側 16:9 廣告位' },
  { thumb: '9', name: '樓盤位置 4', status: '未設定', desc: '樓盤列表右側 9:16 廣告位' },
  { thumb: '10', name: '樓盤位置 5', status: '未設定', desc: '樓盤列表右側 9:16 廣告位' },
];

// 11. 積分流水 mock 資料
const pointRows: PointRow[] = [
  { member: 'patrick', type: '增加', source: 'wallet · ad_reward', points: '50', date: '2026年6月5日' },
  { member: 'patrick', type: '扣除', source: 'property_sale · publish', points: '1,000', date: '2026年5月22日' },
  { member: 'patrick', type: '增加', source: 'wallet · operator_grant', points: '10,000', date: '2026年5月22日' },
];

// 12. ICCTV 大廈與鏡頭 mock 資料
const icctvBuildings: IcctvBuilding[] = [
  { id: 'b1', name: '康睦庭園第二座' },
  { id: 'b2', name: 'Harbour Residence' },
];

const icctvCameras: IcctvCamera[] = [
  { id: 'c1', name: '大堂主鏡頭', location: 'G/F 大堂' },
  { id: 'c2', name: '後門鏡頭', location: '後門出入口' },
];

const icctvSelectedBuilding = ref<string>('');
const icctvSelectedCamera = ref<string>('');
const icctvViewerMeta = computed(() => {
  if (!icctvSelectedCamera.value) return '尚未選擇鏡頭';
  const camera = icctvCameras.find((c) => c.id === icctvSelectedCamera.value);
  return camera ? `${camera.name} · ${camera.location}` : '尚未選擇鏡頭';
});

// 13. 廣告設定選中的素材
const selectedAdSource = ref<string>('');

// 14. 跳轉到新增頁（保留入口，目前無實際路由）
const goCreate = (tab: ManagementTab): void => {
  void router.push({ path: `/account/marketplace/management/${tab}` });
};
</script>

<template>
  <main class="work-shell">
    <aside class="work-sidebar">
      <h1>管理中心</h1>
      <p>管理會員、內容列表、AJO Point 與廣告任務。</p>
      <nav class="work-nav">
        <button
          v-for="item in navItems"
          :key="item.key"
          type="button"
          class="work-nav-item"
          :class="{ on: activeTab === item.key }"
          @click="switchTab(item.key)"
        >
          <AppIcon
            :name="item.icon"
            :size="16"
          />
          <span>{{ item.label }}</span>
        </button>
      </nav>
    </aside>

    <section class="work-main">
      <!-- 家具列表 -->
      <div
        v-if="activeTab === 'admin-furniture'"
        class="work-panel on"
      >
        <section class="staff-list-head">
          <div>
            <div class="staff-kicker">Staff</div>
            <h2 class="staff-title">家具列表</h2>
            <p class="staff-desc">查看二手家私發布狀態，並執行上架、下架與續期。</p>
          </div>
          <button
            type="button"
            class="work-action"
            @click="goCreate('admin-furniture')"
          >
            新增家具
          </button>
        </section>
        <section class="staff-filter-card">
          <div class="staff-filter-row">
            <input
              class="staff-search"
              type="text"
              placeholder="搜尋標題、摘要或編號"
            >
            <select class="staff-select">
              <option>全部狀態</option>
              <option>上架中</option>
              <option>已下架</option>
              <option>已過期</option>
              <option>草稿</option>
            </select>
            <button
              type="button"
              class="work-action"
            >
              搜尋
            </button>
          </div>
        </section>
        <section class="staff-table-wrap">
          <table class="staff-table">
            <thead>
              <tr>
                <th>標題</th>
                <th>發布者</th>
                <th>狀態</th>
                <th>發布時間</th>
                <th>到期時間</th>
                <th>更新時間</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="row in furnitureRows"
                :key="row.code"
              >
                <td>
                  <div class="staff-main-text">{{ row.title }}</div>
                  <div class="staff-sub-text">{{ row.code }}</div>
                </td>
                <td>{{ row.publisher }}</td>
                <td>
                  <span
                    class="staff-status"
                    :class="{ good: row.statusGood }"
                  >{{ row.status }}</span>
                </td>
                <td>{{ row.publishedAt }}</td>
                <td>{{ row.expireAt }}</td>
                <td>{{ row.updatedAt }}</td>
                <td>
                  <div class="staff-actions">
                    <button
                      v-for="action in row.actions"
                      :key="action"
                      type="button"
                      class="staff-action-btn"
                    >{{ action }}</button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
          <div class="staff-pagination">
            <span>第 1-3 筆，共 7 筆</span>
            <div class="staff-page-actions">
              <button
                type="button"
                class="staff-action-btn"
              >上一頁</button>
              <button
                type="button"
                class="staff-action-btn"
              >下一頁</button>
            </div>
          </div>
        </section>
      </div>

      <!-- 樓盤列表 -->
      <div
        v-if="activeTab === 'admin-properties'"
        class="work-panel on"
      >
        <section class="staff-list-head">
          <div>
            <div class="staff-kicker">Staff</div>
            <h2 class="staff-title">樓盤列表</h2>
            <p class="staff-desc">查看樓盤發布狀態，並執行上架、下架、續期與成交管理。</p>
          </div>
          <button
            type="button"
            class="work-action"
            @click="goCreate('admin-properties')"
          >
            新增樓盤
          </button>
        </section>
        <section class="staff-filter-card">
          <div class="staff-filter-row">
            <input
              class="staff-search"
              type="text"
              placeholder="搜尋標題、屋苑、地點"
            >
            <select class="staff-select">
              <option>全部狀態</option>
              <option>上架中</option>
              <option>已下架</option>
              <option>已成交</option>
              <option>草稿</option>
            </select>
            <button
              type="button"
              class="work-action"
            >
              搜尋
            </button>
          </div>
        </section>
        <section class="staff-table-wrap">
          <table class="staff-table">
            <thead>
              <tr>
                <th>標題</th>
                <th>發布者</th>
                <th>狀態</th>
                <th>發布時間</th>
                <th>到期時間</th>
                <th>更新時間</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="row in propertyRows"
                :key="row.code"
              >
                <td>
                  <div class="staff-main-text">{{ row.title }}</div>
                  <div class="staff-sub-text">{{ row.code }}</div>
                </td>
                <td>{{ row.publisher }}</td>
                <td>
                  <span
                    class="staff-status"
                    :class="{ good: row.statusGood }"
                  >{{ row.status }}</span>
                </td>
                <td>{{ row.publishedAt }}</td>
                <td>{{ row.expireAt }}</td>
                <td>{{ row.updatedAt }}</td>
                <td>
                  <div class="staff-actions">
                    <button
                      v-for="action in row.actions"
                      :key="action"
                      type="button"
                      class="staff-action-btn"
                    >{{ action }}</button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
          <div class="staff-pagination">
            <span>第 1-2 筆，共 12 筆</span>
            <div class="staff-page-actions">
              <button
                type="button"
                class="staff-action-btn"
              >上一頁</button>
              <button
                type="button"
                class="staff-action-btn"
              >下一頁</button>
            </div>
          </div>
        </section>
      </div>

      <!-- 住宅列表 -->
      <div
        v-if="activeTab === 'admin-homes'"
        class="work-panel on"
      >
        <section class="staff-list-head">
          <div>
            <div class="staff-kicker">Residential</div>
            <h2 class="staff-title">住宅列表</h2>
            <p class="staff-desc">管理服務式住宅、住戶綁定與單位狀態。</p>
          </div>
          <button
            type="button"
            class="work-action"
            @click="goCreate('admin-homes')"
          >
            新增住宅
          </button>
        </section>
        <section class="staff-filter-card">
          <div class="staff-filter-row">
            <input
              class="staff-search"
              type="text"
              placeholder="搜尋住宅、單位或住戶"
            >
            <select class="staff-select">
              <option>全部狀態</option>
              <option>已綁定</option>
              <option>待確認</option>
              <option>已停用</option>
            </select>
            <button
              type="button"
              class="work-action"
            >
              搜尋
            </button>
          </div>
        </section>
        <section class="staff-table-wrap">
          <table class="staff-table">
            <thead>
              <tr>
                <th>住宅</th>
                <th>單位</th>
                <th>住戶</th>
                <th>狀態</th>
                <th>更新時間</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="row in homeRows"
                :key="row.name + row.unit"
              >
                <td>
                  <div class="staff-main-text">{{ row.name }}</div>
                  <div class="staff-sub-text">{{ row.sub }}</div>
                </td>
                <td>{{ row.unit }}</td>
                <td>{{ row.resident }}</td>
                <td>
                  <span
                    class="staff-status"
                    :class="{ good: row.statusGood }"
                  >{{ row.status }}</span>
                </td>
                <td>{{ row.updatedAt }}</td>
                <td>
                  <div class="staff-actions">
                    <button
                      v-for="action in row.actions"
                      :key="action"
                      type="button"
                      class="staff-action-btn"
                    >{{ action }}</button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
          <div class="staff-pagination">
            <span>第 1-2 筆，共 12 筆</span>
            <div class="staff-page-actions">
              <button
                type="button"
                class="staff-action-btn"
              >上一頁</button>
              <button
                type="button"
                class="staff-action-btn"
              >下一頁</button>
            </div>
          </div>
        </section>
      </div>

      <!-- ICCTV -->
      <div
        v-if="activeTab === 'admin-icctv'"
        class="work-panel on"
      >
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
        <section class="icctv-manage-grid">
          <div class="work-card">
            <div class="icctv-building-select">
              <select
                v-model="icctvSelectedBuilding"
                class="staff-select"
              >
                <option value="">選擇大廈</option>
                <option
                  v-for="b in icctvBuildings"
                  :key="b.id"
                  :value="b.id"
                >{{ b.name }}</option>
              </select>
              <button
                type="button"
                class="work-mini-btn"
              >重新載入</button>
            </div>
            <div class="icctv-camera-list">
              <button
                v-for="c in icctvCameras"
                :key="c.id"
                type="button"
                class="icctv-camera-btn"
                :class="{ on: icctvSelectedCamera === c.id }"
                @click="icctvSelectedCamera = c.id"
              >
                <span>
                  {{ c.name }}
                  <small>{{ c.location }}</small>
                </span>
              </button>
            </div>
            <div class="icctv-viewer-bar">
              <div class="icctv-viewer-meta">{{ icctvViewerMeta }}</div>
              <div class="icctv-viewer-actions">
                <button
                  type="button"
                  class="work-mini-btn"
                >新窗口</button>
              </div>
            </div>
            <div class="icctv-viewer">
              <div
                v-if="!icctvSelectedCamera"
                class="icctv-empty"
              >
                <strong>未選擇鏡頭</strong>
                <span>選擇大廈與鏡頭後，監控畫面會顯示在此處。</span>
              </div>
              <div
                v-else
                class="icctv-viewer-placeholder"
              >
                <span>{{ icctvViewerMeta }}</span>
              </div>
            </div>
          </div>
          <aside class="icctv-manage-side">
            <div class="icctv-manage-panel">
              <h4>接入狀態</h4>
              <div class="icctv-tag-row">
                <span class="icctv-tag on">Orange Pi</span>
                <span class="icctv-tag">WebRTC</span>
                <span class="icctv-tag">iCCTV</span>
              </div>
              <p class="icctv-note">Staff 先選擇大廈，再選擇鏡頭。畫面使用後台返回的監控 URL 嵌入。</p>
            </div>
            <div class="icctv-manage-panel">
              <h4>嵌入方式</h4>
              <div class="icctv-note">目前使用 iframe 預覽。若來源頁限制嵌入或瀏覽器阻擋混合內容，可使用新窗口開啟，後續再接專用 WebRTC 播放器。</div>
            </div>
          </aside>
        </section>
      </div>

      <!-- 會員列表 -->
      <div
        v-if="activeTab === 'admin-members'"
        class="work-panel on"
      >
        <section class="staff-list-head">
          <div>
            <div class="staff-kicker">Members</div>
            <h2 class="staff-title">會員列表</h2>
            <p class="staff-desc">查看會員身份、狀態、角色與資料來源。</p>
          </div>
          <button
            type="button"
            class="work-action"
            @click="goCreate('admin-members')"
          >
            新增會員
          </button>
        </section>
        <section class="staff-filter-card">
          <div class="staff-filter-row">
            <input
              class="staff-search"
              type="text"
              placeholder="搜尋姓名、電郵或手機"
            >
            <select class="staff-select">
              <option>全部狀態</option>
              <option>active</option>
              <option>pending</option>
              <option>disabled</option>
            </select>
            <button
              type="button"
              class="work-action"
            >
              搜尋
            </button>
          </div>
        </section>
        <section class="staff-table-wrap">
          <table class="staff-table">
            <thead>
              <tr>
                <th>會員</th>
                <th>會員類型</th>
                <th>主要角色</th>
                <th>狀態</th>
                <th>資料來源</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="row in memberRows"
                :key="row.name + row.sub"
              >
                <td>
                  <div class="staff-main-text">{{ row.name }}</div>
                  <div class="staff-sub-text">{{ row.sub }}</div>
                </td>
                <td>{{ row.type }}</td>
                <td>{{ row.role }}</td>
                <td>
                  <span
                    class="staff-status"
                    :class="{ good: row.statusGood }"
                  >{{ row.status }}</span>
                </td>
                <td>{{ row.source }}</td>
                <td>
                  <div class="staff-actions">
                    <button
                      v-for="action in row.actions"
                      :key="action"
                      type="button"
                      class="staff-action-btn"
                    >{{ action }}</button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
          <div class="staff-pagination">
            <span>第 1-2 筆，共 48 筆</span>
            <div class="staff-page-actions">
              <button
                type="button"
                class="staff-action-btn"
              >上一頁</button>
              <button
                type="button"
                class="staff-action-btn"
              >下一頁</button>
            </div>
          </div>
        </section>
      </div>

      <!-- 廣告列表 -->
      <div
        v-if="activeTab === 'admin-ads'"
        class="work-panel on"
      >
        <section class="staff-list-head">
          <div>
            <div class="staff-kicker">Staff</div>
            <h2 class="staff-title">廣告列表</h2>
            <p class="staff-desc">查看廣告素材狀態，並執行上架、下架與續期。</p>
          </div>
          <button
            type="button"
            class="work-action"
            @click="goCreate('admin-ads')"
          >
            新增廣告
          </button>
        </section>
        <section class="staff-filter-card">
          <div class="staff-filter-row">
            <input
              class="staff-search"
              type="text"
              placeholder="搜尋廣告標題、位置或編號"
            >
            <select class="staff-select">
              <option>全部狀態</option>
              <option>投放中</option>
              <option>排期中</option>
              <option>草稿</option>
              <option>已下架</option>
            </select>
            <button
              type="button"
              class="work-action"
            >
              搜尋
            </button>
          </div>
        </section>
        <section class="staff-table-wrap">
          <table class="staff-table">
            <thead>
              <tr>
                <th>標題</th>
                <th>發布者</th>
                <th>狀態</th>
                <th>發布時間</th>
                <th>到期時間</th>
                <th>更新時間</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="row in adRows"
                :key="row.code"
              >
                <td>
                  <div class="staff-main-text">{{ row.title }}</div>
                  <div class="staff-sub-text">{{ row.code }}</div>
                </td>
                <td>{{ row.publisher }}</td>
                <td>
                  <span
                    class="staff-status"
                    :class="{ good: row.statusGood }"
                  >{{ row.status }}</span>
                </td>
                <td>{{ row.publishedAt }}</td>
                <td>{{ row.expireAt }}</td>
                <td>{{ row.updatedAt }}</td>
                <td>
                  <div class="staff-actions">
                    <button
                      v-for="action in row.actions"
                      :key="action"
                      type="button"
                      class="staff-action-btn"
                    >{{ action }}</button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
          <div class="staff-pagination">
            <span>第 1-2 筆，共 8 筆</span>
            <div class="staff-page-actions">
              <button
                type="button"
                class="staff-action-btn"
              >上一頁</button>
              <button
                type="button"
                class="staff-action-btn"
              >下一頁</button>
            </div>
          </div>
        </section>
      </div>

      <!-- 廣告設定 -->
      <div
        v-if="activeTab === 'admin-ad-settings'"
        class="work-panel on"
      >
        <section class="work-hero">
          <div>
            <div class="work-kicker">Ad Settings</div>
            <h2 class="work-title">廣告設定</h2>
            <p class="work-desc">左側選擇廣告素材，右側放入家具和樓盤等頁面的廣告位置。</p>
          </div>
          <button
            type="button"
            class="work-action"
          >
            儲存
          </button>
        </section>
        <section class="admin-ad-settings">
          <div class="work-card">
            <div class="work-card-title">廣告素材</div>
            <div class="ad-source-list">
              <div
                v-for="item in adSources"
                :key="item.thumb"
                class="ad-source-item"
                :class="{ on: selectedAdSource === item.thumb }"
                @click="selectedAdSource = item.thumb"
              >
                <div class="ad-source-thumb">{{ item.thumb }}</div>
                <div>
                  <div class="ad-source-title">{{ item.title }}</div>
                  <div class="ad-source-meta">{{ item.meta }}</div>
                  <button
                    type="button"
                    class="work-mini-btn"
                  >選擇</button>
                </div>
              </div>
            </div>
          </div>
          <div class="work-card">
            <div class="work-card-title">廣告位</div>
            <div class="ad-slot-grid">
              <div
                v-for="item in adSlots"
                :key="item.thumb"
                class="ad-slot-card"
              >
                <div class="ad-slot-thumb">{{ item.thumb }}</div>
                <div>
                  <div class="ad-slot-name">{{ item.name }}</div>
                  <div class="ad-slot-status">{{ item.status }}</div>
                  <div class="ad-slot-desc">{{ item.desc }}</div>
                  <div class="ad-slot-actions">
                    <button
                      type="button"
                      class="work-mini-btn"
                    >放入選擇廣告</button>
                    <button
                      type="button"
                      class="work-mini-btn"
                    >清空</button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>
      </div>

      <!-- 積分流水 -->
      <div
        v-if="activeTab === 'admin-points'"
        class="work-panel on"
      >
        <section class="staff-list-head">
          <div>
            <div class="staff-kicker">Points</div>
            <h2 class="staff-title">積分流水</h2>
            <p class="staff-desc">查看 AJO Point 派發、扣除和廣告獎勵記錄。</p>
          </div>
        </section>
        <section class="staff-filter-card">
          <div class="staff-filter-row">
            <input
              class="staff-search"
              type="text"
              placeholder="搜尋會員或來源"
            >
            <select class="staff-select">
              <option>全部類型</option>
              <option>增加</option>
              <option>扣除</option>
            </select>
            <button
              type="button"
              class="work-action"
            >
              搜尋
            </button>
          </div>
        </section>
        <section class="staff-table-wrap">
          <table class="staff-table">
            <thead>
              <tr>
                <th>會員</th>
                <th>類型</th>
                <th>來源</th>
                <th>積分</th>
                <th>日期</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="(row, idx) in pointRows"
                :key="idx"
              >
                <td>
                  <div class="staff-main-text">{{ row.member }}</div>
                </td>
                <td>{{ row.type }}</td>
                <td>{{ row.source }}</td>
                <td>{{ row.points }}</td>
                <td>{{ row.date }}</td>
              </tr>
            </tbody>
          </table>
        </section>
      </div>
    </section>
  </main>
</template>

<style scoped>
/* 1. 整體殼層：左側導航 + 右側內容 */
.work-shell {
  display: grid;
  grid-template-columns: 240px minmax(0, 1fr);
  gap: 16px;
  width: 100%;
  max-width: 1180px;
  margin: 0 auto;
  padding: 12px var(--layout-page-padding-inline) 16px;
  color: rgb(var(--color-text));
}

/* 2. 左側導航 */
.work-sidebar {
  position: sticky;
  top: calc(var(--app-header-offset, 0rem) + 1rem);
  align-self: start;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 16px;
}

.work-sidebar h1 {
  margin: 0;
  color: rgb(var(--color-primary));
  font-family: var(--font-display);
  font-size: 26px;
  font-weight: 400;
  line-height: 1.2;
}

.work-sidebar p {
  margin: 8px 0 0;
  color: rgb(var(--color-ink-3));
  font-size: 12px;
  line-height: 1.7;
}

.work-nav {
  display: grid;
  gap: 2px;
  margin-top: 18px;
}

.work-nav-item {
  position: relative;
  display: flex;
  align-items: center;
  gap: 8px;
  min-height: 34px;
  border: 0;
  border-radius: 0;
  background: transparent;
  color: rgb(var(--color-ink-2));
  cursor: pointer;
  font-family: inherit;
  font-size: 13px;
  font-weight: 500;
  padding: 8px 2px;
  text-align: left;
  transition:
    background-color 0.15s ease,
    color 0.15s ease;
}

.work-nav-item::after {
  content: '';
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  height: 2px;
  background: rgb(var(--color-primary));
  transform: scaleX(0);
  transform-origin: left center;
  transition: transform 0.24s ease;
}

.work-nav-item:hover {
  background: rgb(var(--color-primary-soft));
  color: rgb(var(--color-text));
}

.work-nav-item.on {
  background: transparent;
  color: rgb(var(--color-primary));
  font-weight: 700;
  outline: none;
}

.work-nav-item.on:hover {
  background: rgb(var(--color-primary-soft));
}

.work-nav-item.on::after {
  transform: scaleX(1);
}

/* 3. 右側主內容 */
.work-main {
  display: grid;
  gap: 12px;
  min-width: 0;
  align-content: start;
}

.work-panel.on {
  display: grid;
  gap: 14px;
  align-content: start;
}

/* 4. 列表標題區 */
.staff-list-head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 12px;
}

.staff-kicker {
  color: rgb(var(--color-ink-3));
  font-size: 10px;
  font-weight: 800;
  letter-spacing: 1.6px;
  text-transform: uppercase;
}

.staff-title {
  margin: 6px 0 0;
  color: rgb(var(--color-text));
  font-family: var(--font-display);
  font-size: 30px;
  font-weight: 400;
  line-height: 1.1;
}

.staff-desc {
  margin: 6px 0 0;
  color: rgb(var(--color-ink-3));
  font-size: 13px;
  font-weight: 600;
  line-height: 1.5;
}

/* 5. 篩選區 */
.staff-filter-card {
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 8px 10px;
  align-self: start;
}

.staff-filter-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 180px 76px;
  gap: 10px;
  align-items: center;
}

.staff-search,
.staff-select {
  height: 34px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 6px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
  font-family: inherit;
  font-size: 13px;
  font-weight: 700;
  padding: 0 12px;
}

.staff-filter-row .work-action {
  height: 34px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0 12px;
  font-size: 12px;
}

/* 6. 表格區 */
.staff-table-wrap {
  margin-top: 0;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  overflow: hidden;
}

.staff-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 14px;
}

.staff-table th {
  border-bottom: 1px solid rgb(var(--color-border));
  color: rgb(var(--color-ink-3));
  font-size: 13px;
  font-weight: 800;
  padding: 14px 18px;
  text-align: left;
}

.staff-table td {
  border-bottom: 1px solid rgb(var(--color-surface-3));
  color: rgb(var(--color-ink-2));
  padding: 16px 18px;
  vertical-align: middle;
}

.staff-table tr:last-child td {
  border-bottom: 0;
}

.staff-main-text {
  color: rgb(var(--color-text));
  font-size: 15px;
  font-weight: 800;
}

.staff-sub-text {
  margin-top: 4px;
  color: rgb(var(--color-ink-3));
  font-size: 12px;
  font-weight: 600;
  max-width: 190px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.staff-status {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  background: rgb(var(--color-primary-soft));
  color: rgb(var(--color-primary));
  font-size: 12px;
  font-weight: 800;
  padding: 6px 10px;
  white-space: nowrap;
}

.staff-status.good {
  background: rgb(var(--color-success-bg));
  color: rgb(var(--color-success));
}

.staff-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.staff-action-btn {
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
  cursor: pointer;
  font-family: inherit;
  font-size: 13px;
  font-weight: 800;
  padding: 8px 13px;
}

.staff-pagination {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-top: 1px solid rgb(var(--color-border));
  padding: 12px 18px;
  color: rgb(var(--color-ink-3));
  font-size: 13px;
  font-weight: 700;
}

.staff-page-actions {
  display: flex;
  gap: 8px;
}

/* 7. 主要操作按鈕 */
.work-action {
  border: 0;
  border-radius: 6px;
  background: rgb(var(--color-primary));
  color: #fff;
  cursor: pointer;
  font-family: inherit;
  font-size: 12px;
  font-weight: 600;
  padding: 10px 14px;
  white-space: nowrap;
}

/* 8. ICCTV 區 */
.icctv-manage-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 320px;
  gap: 12px;
  align-items: start;
}

.work-card {
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 16px;
}

.work-card-title {
  margin-bottom: 10px;
  color: rgb(var(--color-text));
  font-size: 14px;
  font-weight: 600;
}

.icctv-building-select {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
}

.icctv-building-select .staff-select {
  flex: 1;
}

.icctv-building-select .work-mini-btn {
  white-space: nowrap;
}

.icctv-camera-list {
  display: grid;
  gap: 8px;
  width: 100%;
  margin-top: 12px;
}

.icctv-camera-btn {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
  cursor: pointer;
  font-family: inherit;
  font-size: 13px;
  font-weight: 800;
  padding: 10px 12px;
  text-align: left;
}

.icctv-camera-btn.on {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary-soft));
  color: rgb(var(--color-primary));
}

.icctv-camera-btn small {
  display: block;
  color: rgb(var(--color-ink-3));
  font-size: 11px;
  font-weight: 700;
}

.icctv-viewer-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  padding: 12px 0 0;
}

.icctv-viewer-meta {
  color: rgb(var(--color-ink-3));
  font-size: 12px;
  font-weight: 700;
}

.icctv-viewer-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.icctv-viewer {
  position: relative;
  min-height: 520px;
  margin-top: 8px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: #111827;
  overflow: hidden;
}

.icctv-viewer-placeholder {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #e5e7eb;
  font-size: 14px;
  font-weight: 700;
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
  color: #e5e7eb;
  text-align: center;
  padding: 20px;
}

.icctv-empty strong {
  font-size: 18px;
  font-weight: 800;
}

.icctv-empty span {
  font-size: 13px;
  color: #94a3b8;
  font-weight: 600;
  max-width: 280px;
  line-height: 1.6;
}

.icctv-manage-side {
  display: grid;
  gap: 12px;
}

.icctv-manage-panel {
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 14px;
}

.icctv-manage-panel h4 {
  margin: 0 0 10px;
  color: rgb(var(--color-text));
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
  background: rgb(var(--color-surface-2));
  color: rgb(var(--color-ink-2));
  font-size: 11px;
  font-weight: 800;
  padding: 5px 9px;
}

.icctv-tag.on {
  background: rgb(var(--color-primary-soft));
  color: rgb(var(--color-primary));
}

.icctv-note {
  margin-top: 12px;
  color: rgb(var(--color-ink-3));
  font-size: 12px;
  line-height: 1.6;
}

/* 9. 小型按鈕 */
.work-mini-btn {
  border: 1px solid rgb(var(--color-border));
  border-radius: 6px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
  cursor: pointer;
  font-family: inherit;
  font-size: 12px;
  font-weight: 600;
  padding: 7px 10px;
}

/* 10. 廣告設定區 */
.work-hero {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 18px;
  border: 0;
  border-bottom: 1px solid rgb(var(--color-border));
  border-radius: 0;
  background: transparent;
  margin: 0;
  padding: 0 0 10px;
}

.work-kicker {
  margin-bottom: 4px;
  color: rgb(var(--color-ink-3));
  font-size: 9px;
  letter-spacing: 1.4px;
  text-transform: uppercase;
}

.work-title {
  margin: 0;
  color: rgb(var(--color-text));
  font-size: 20px;
  font-weight: 600;
  line-height: 1.25;
}

.work-desc {
  max-width: 560px;
  margin: 5px 0 0;
  color: rgb(var(--color-ink-3));
  font-size: 13px;
  line-height: 1.5;
}

.admin-ad-settings {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1.4fr);
  gap: 12px;
  align-items: start;
}

.ad-source-list {
  display: grid;
  gap: 10px;
  max-height: 620px;
  overflow: auto;
  padding-right: 4px;
}

.ad-source-item {
  display: grid;
  grid-template-columns: 92px minmax(0, 1fr);
  gap: 12px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 10px;
  cursor: pointer;
}

.ad-source-item.on {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary-soft));
}

.ad-source-thumb {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 72px;
  border-radius: 6px;
  background: rgb(var(--color-surface-3));
  color: rgb(var(--color-ink-3));
  font-size: 24px;
  font-weight: 700;
}

.ad-source-title {
  color: rgb(var(--color-text));
  font-size: 13px;
  font-weight: 700;
}

.ad-source-meta {
  margin-top: 4px;
  color: rgb(var(--color-ink-3));
  font-size: 11px;
  line-height: 1.5;
}

.ad-slot-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.ad-slot-card {
  display: grid;
  grid-template-columns: 92px minmax(0, 1fr);
  gap: 12px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 12px;
}

.ad-slot-thumb {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 96px;
  border-radius: 6px;
  background: rgb(var(--color-surface-3));
  color: rgb(var(--color-ink-3));
  font-size: 28px;
  font-weight: 500;
}

.ad-slot-name {
  color: rgb(var(--color-ink-3));
  font-size: 12px;
  font-weight: 700;
}

.ad-slot-status {
  margin-top: 3px;
  color: rgb(var(--color-text));
  font-size: 14px;
  font-weight: 800;
}

.ad-slot-desc {
  margin-top: 6px;
  color: rgb(var(--color-ink-3));
  font-size: 12px;
  line-height: 1.5;
}

.ad-slot-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-top: 10px;
}

/* 11. 響應式：平板 */
@media (max-width: 1023px) {
  .work-shell {
    grid-template-columns: 1fr;
    padding: 14px var(--layout-page-padding-inline);
  }

  .work-sidebar {
    position: static;
  }

  .work-nav {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .icctv-manage-grid {
    grid-template-columns: 1fr;
  }

  .admin-ad-settings {
    grid-template-columns: 1fr;
  }
}

/* 12. 響應式：手機 */
@media (max-width: 647px) {
  .staff-filter-row {
    grid-template-columns: 1fr;
  }

  .work-nav {
    grid-template-columns: 1fr;
  }

  .work-hero {
    align-items: flex-start;
    flex-direction: column;
  }

  .ad-slot-grid {
    grid-template-columns: 1fr;
  }

  .staff-list-head {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
