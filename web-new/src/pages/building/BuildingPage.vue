<!--
 * 我的大廈頁。
 * 1. 高保真還原 HTML 設計稿 page-affairs 雙欄布局（左側大廈導航 + 右側面板）。
 * 2. 提供最新通告、大廈資料、大廈財務、業戶帳目、申請表格、意見提供、智能門禁、視像監控、設備監測九個面板。
 * 3. 全部使用靜態 mock 資料，不接入任何 API。
-->
<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';

type AffairsTab =
  | 'affairs-notices'
  | 'affairs-building'
  | 'affairs-finance'
  | 'affairs-owner-account'
  | 'affairs-forms'
  | 'affairs-feedback'
  | 'affairs-access'
  | 'affairs-icctv'
  | 'affairs-equipment';

interface NavItem {
  target: AffairsTab;
  label: string;
}

interface NoticeRow {
  code: string;
  title: string;
  type: string;
  publishDate: string;
  expireDate: string;
}

interface BuildingField {
  label: string;
  value: string;
}

interface BuildingDocCard {
  title: string;
  desc: string;
  count: string;
  empty: string;
}

interface FloorPlanRow {
  title: string;
  date: string;
  month: string;
}

interface FinanceOverRow {
  unit: string;
  m11: string;
  m10: string;
  m09: string;
  mBefore: string;
  m11Status: 'paid' | 'due';
  m10Status: 'paid' | 'due';
  m09Status: 'paid' | 'due';
  mBeforeStatus: 'paid' | 'due';
}

interface FinanceReportRow {
  date: string;
  title: string;
}

interface AuditRow {
  date: string;
  item: string;
}

interface FormRow {
  title: string;
}

interface FeedbackRecord {
  subject: string;
  category: string;
  status: 'processing' | 'done';
  statusText: string;
  updatedAt: string;
}

interface EquipmentRow {
  device: string;
  location: string;
  status: 'good' | 'warn';
  statusText: string;
  updatedAt: string;
}

const router = useRouter();

// 1. 左側導航項目
const navItems: NavItem[] = [
  { target: 'affairs-notices', label: '最新通告' },
  { target: 'affairs-building', label: '大廈資料' },
  { target: 'affairs-finance', label: '大廈財務' },
  { target: 'affairs-owner-account', label: '業戶帳目' },
  { target: 'affairs-forms', label: '申請表格' },
  { target: 'affairs-feedback', label: '意見提供' },
  { target: 'affairs-access', label: '智能門禁' },
  { target: 'affairs-icctv', label: '視像監控' },
  { target: 'affairs-equipment', label: '設備監測' },
];

// 2. 當前啟用面板
const activeTab = ref<AffairsTab>('affairs-notices');

// 3. 大廈財務子面板
type FinanceSubTab = 'acct-overview' | 'acct-finance' | 'acct-audit' | 'acct-payment-records';
const financeSubTab = ref<FinanceSubTab>('acct-overview');

// 4. 最新通告 mock 資料
const noticeBuildings = ['仁美大廈', '康睦庭園第二座'];
const selectedNoticeBuilding = ref('仁美大廈');
const notices: NoticeRow[] = [
  {
    code: 'bulk/0145100_90c1340f10e849bd9c69',
    title: '清潔員大年初一休假一天',
    type: '一般',
    publishDate: 'Jan. 24, 2025',
    expireDate: 'Jan. 31, 2025',
  },
  {
    code: 'bulk/0145100_91b2451f21e950ce8d70',
    title: '大廈外牆維修工程通知',
    type: '工程',
    publishDate: 'Jun. 10, 2026',
    expireDate: 'Jul. 10, 2026',
  },
  {
    code: 'bulk/0145100_92c3562f32fa61fa9e81',
    title: '管理費季度繳款提醒',
    type: '財務',
    publishDate: 'Jun. 15, 2026',
    expireDate: 'Jul. 15, 2026',
  },
];

// 5. 大廈資料 mock 資料
const buildingFields: BuildingField[] = [
  { label: '落成年份', value: '1991' },
  { label: '樓層總數', value: '15' },
  { label: '單位總數', value: '32' },
  { label: '車位總數', value: '0' },
  { label: '法團名稱', value: '時安大廈(洋松街)業主立案法團' },
  { label: '管理處電話', value: '2393 4230' },
  { label: '管理公司名稱', value: '顯安居物業管理股份有限公司' },
  { label: '管理公司電話', value: '2384 2251' },
  { label: '管理公司電郵', value: 'info@showsecurity.com.hk' },
  { label: '管理公司傳真', value: '2384 2243' },
  { label: '民政事務處電話', value: '油尖旺: 2399 2111' },
  { label: '資料來源', value: 'iSmart 基本資料' },
];

const buildingDocCards: BuildingDocCard[] = [
  {
    title: '表格',
    desc: '住戶常用或職員常用的基本表格文件。',
    count: '0 份',
    empty: '目前未有表格，可按右上角按鈕新增。',
  },
  {
    title: '大廈資訊',
    desc: '對外發佈或內部參考的大廈介紹與基本資訊附件。',
    count: '0 份',
    empty: '目前未有大廈資訊，可按右上角按鈕新增。',
  },
  {
    title: '平面圖',
    desc: '平面圖、設施位置圖及相關圖則文件。',
    count: '4 份',
    empty: '已匯入舊系統平面圖資料。',
  },
];

const floorPlans: FloorPlanRow[] = [
  { title: '15TH.FL.PLAN', date: '-', month: '-' },
  { title: '2ND-14TH,FL.PLAN', date: '-', month: '-' },
  { title: '1ST.FLOOR PLAN', date: '-', month: '-' },
  { title: 'GROUND FL.PLAN', date: '-', month: '-' },
];

// 6. 大廈財務 mock 資料
const financeOverview: FinanceOverRow[] = [
  { unit: 'G樓 01', m11: '-463.0', m10: '已付', m09: '已付', mBefore: '已付', m11Status: 'due', m10Status: 'paid', m09Status: 'paid', mBeforeStatus: 'paid' },
  { unit: 'G樓 02', m11: '-463.0', m10: '已付', m09: '已付', mBefore: '已付', m11Status: 'due', m10Status: 'paid', m09Status: 'paid', mBeforeStatus: 'paid' },
  { unit: 'G樓 03', m11: '-463.0', m10: '已付', m09: '已付', mBefore: '已付', m11Status: 'due', m10Status: 'paid', m09Status: 'paid', mBeforeStatus: 'paid' },
  { unit: 'G樓 04', m11: '-463.0', m10: '已付', m09: '已付', mBefore: '已付', m11Status: 'due', m10Status: 'paid', m09Status: 'paid', mBeforeStatus: 'paid' },
  { unit: 'G樓 5A', m11: '已付', m10: '已付', m09: '已付', mBefore: '已付', m11Status: 'paid', m10Status: 'paid', m09Status: 'paid', mBeforeStatus: 'paid' },
  { unit: 'G樓 5B', m11: '-543.0', m10: '已付', m09: '已付', mBefore: '已付', m11Status: 'due', m10Status: 'paid', m09Status: 'paid', mBeforeStatus: 'paid' },
  { unit: '01樓 A', m11: '-476.0', m10: '已付', m09: '已付', mBefore: '已付', m11Status: 'due', m10Status: 'paid', m09Status: 'paid', mBeforeStatus: 'paid' },
  { unit: '01樓 B', m11: '-476.0', m10: '已付', m09: '已付', mBefore: '已付', m11Status: 'due', m10Status: 'paid', m09Status: 'paid', mBeforeStatus: 'paid' },
  { unit: '01樓 E', m11: '-1427.0', m10: '已付', m09: '已付', mBefore: '已付', m11Status: 'due', m10Status: 'paid', m09Status: 'paid', mBeforeStatus: 'paid' },
  { unit: '02樓 B', m11: '-476.0', m10: '-476.0', m09: '已付', mBefore: '已付', m11Status: 'due', m10Status: 'due', m09Status: 'paid', mBeforeStatus: 'paid' },
  { unit: '02樓 E', m11: '-1427.0', m10: '-1427.0', m09: '-1427.0', mBefore: '-1427.0', m11Status: 'due', m10Status: 'due', m09Status: 'due', mBeforeStatus: 'due' },
  { unit: '03樓 A', m11: '-619.0', m10: '-619.0', m09: '已付', mBefore: '已付', m11Status: 'due', m10Status: 'due', m09Status: 'paid', mBeforeStatus: 'paid' },
  { unit: '07樓 D', m11: '-1006.0', m10: '-1006.0', m09: '-1006.0', mBefore: '-5030.0', m11Status: 'due', m10Status: 'due', m09Status: 'due', mBeforeStatus: 'due' },
  { unit: '15樓 B', m11: '-1006.0', m10: '-1006.0', m09: '-1006.0', mBefore: '-1006.0', m11Status: 'due', m10Status: 'due', m09Status: 'due', mBeforeStatus: 'due' },
  { unit: '22樓 B', m11: '-1006.0', m10: '已付', m09: '已付', mBefore: '已付', m11Status: 'due', m10Status: 'paid', m09Status: 'paid', mBeforeStatus: 'paid' },
];

const financeReports: FinanceReportRow[] = [
  { date: '2027-07', title: 'YIG 財務報告 2022-07' },
  { date: '2023-03', title: 'YIG 財務報告 2023-03' },
  { date: '2023-02', title: 'YIG 財務報告 2023-02' },
  { date: '2022-12', title: 'YIG 財務報告 2022-12' },
  { date: '2022-11', title: 'YIG 財務報告 2022-11' },
];

const auditReports: AuditRow[] = [
  { date: '2021-11', item: 'YIG 財務報告 2021-11' },
  { date: '2018', item: '仁英大廈核數報告 2018' },
  { date: '2017', item: '仁英大廈核數報告 2017' },
  { date: '2016', item: '仁英大廈核數報告 2016' },
  { date: '2015', item: '仁英大廈核數報告 2015' },
];

// 7. 申請表格 mock 資料
const formBuildings = ['時安大廈', '康睦庭園第二座'];
const selectedFormBuilding = ref('時安大廈');
const forms: FormRow[] = [
  { title: '單位裝修申請表' },
  { title: '暫停食水、沖廁水申請表' },
  { title: '申請增設電錶、加大配電資料記錄跟進表' },
];

// 8. 意見提供 mock 資料
const feedbackBuildings = [
  '仁英大廈', '協和大廈', '華興大廈(269號)', '華興大廈(271號)', '豐富大廈',
  '得運大廈', '天富大廈', '新萬利大廈', '萬高大廈B座', '時安大廈',
  '東昇樓', '華園', '榮森工業第二大廈', '康睦庭園第二座', '南昌苑',
];
const feedbackTitles = [
  '門卡報失', '冷氣滴水', '噪音滋擾', '樓梯雜物', '水質問題',
  '渠務問題', '保安事宜', '清潔衛生', '增加服務', '電力問題', '其他事宜',
];
const selectedFeedbackBuilding = ref('協和大廈');
const selectedFeedbackTitle = ref('');
const feedbackContent = ref('');
const feedbackRecords: FeedbackRecord[] = [
  { subject: '公共走廊照明檢查', category: '維修', status: 'processing', statusText: '處理中', updatedAt: '今天 10:20' },
  { subject: '大堂清潔建議', category: '清潔', status: 'done', statusText: '已完成', updatedAt: '昨天 16:45' },
];

// 9. 設備監測 mock 資料
const equipmentRows: EquipmentRow[] = [
  { device: '升降機 1 號', location: '大堂', status: 'good', statusText: '正常', updatedAt: '今天 10:20' },
  { device: '水泵房', location: '地庫', status: 'warn', statusText: '需檢查', updatedAt: '今天 09:40' },
  { device: '照明系統', location: '公共走廊', status: 'good', statusText: '正常', updatedAt: '昨天 18:10' },
];

// 10. 切換主面板
const switchTab = (target: AffairsTab) => {
  activeTab.value = target;
};

// 11. 切換財務子面板
const switchFinanceSub = (target: FinanceSubTab) => {
  financeSubTab.value = target;
};

// 12. 提交意見回饋
const submitFeedback = () => {
  feedbackContent.value = '';
};

// 13. 重新整理通告
const refreshNotices = () => {
  // 靜態 mock，無需操作
};

// 14. 跳轉至大廈詳情
const goBuildingDetail = () => {
  router.push('/building/detail');
};
</script>

<template>
  <div class="page-affairs">
    <div class="work-shell">
      <!-- 左側大廈導航 -->
      <aside class="work-sidebar">
        <h1>我的大廈</h1>
        <nav class="work-nav">
          <button
            v-for="item in navItems"
            :key="item.target"
            type="button"
            class="work-nav-item"
            :class="{ on: activeTab === item.target }"
            @click="switchTab(item.target)"
          >
            {{ item.label }}
          </button>
        </nav>
      </aside>

      <!-- 右側主內容 -->
      <main class="work-main">
        <!-- 最新通告 -->
        <div
          v-show="activeTab === 'affairs-notices'"
          class="work-panel on"
        >
          <section class="work-hero">
            <div>
              <div class="work-kicker">Building Notices</div>
              <h2 class="work-title">最新通告</h2>
            </div>
          </section>
          <section class="notice-admin-grid">
            <div class="notice-admin-card">
              <h3>大廈選擇</h3>
              <label class="notice-admin-label">選擇大廈:</label>
              <select
                v-model="selectedNoticeBuilding"
                class="notice-select"
              >
                <option
                  v-for="b in noticeBuildings"
                  :key="b"
                  :value="b"
                >
                  {{ b }}
                </option>
              </select>
            </div>
          </section>
          <section class="work-card">
            <div class="notice-current">
              <span>目前顯示: 所有通告</span>
              <button
                type="button"
                class="notice-action-btn"
                @click="refreshNotices"
              >
                重新整理
              </button>
            </div>
            <table class="work-table notice-table">
              <thead>
                <tr>
                  <th>編號</th>
                  <th>標題</th>
                  <th>類型</th>
                  <th>發佈日期</th>
                  <th>下架日期</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="n in notices"
                  :key="n.code"
                >
                  <td>{{ n.code }}</td>
                  <td>{{ n.title }}</td>
                  <td>{{ n.type }}</td>
                  <td>{{ n.publishDate }}</td>
                  <td>{{ n.expireDate }}</td>
                  <td>
                    <div class="notice-table-actions">
                      <button
                        type="button"
                        class="notice-action-btn"
                      >
                        查閱
                      </button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </section>
        </div>

        <!-- 大廈資料 -->
        <div
          v-show="activeTab === 'affairs-building'"
          class="work-panel on"
        >
          <section class="work-hero">
            <div>
              <div class="work-kicker">Building Profile</div>
              <h2 class="work-title">大廈基本資料管理</h2>
              <p class="work-desc">在此維護大廈基本欄位，並集中管理表格、大廈資訊附件及平面圖。</p>
            </div>
            <button
              type="button"
              class="work-action"
            >
              儲存資料
            </button>
          </section>
          <section class="building-summary-grid">
            <div class="work-card">
              <div class="work-card-title">機構</div>
              <div class="work-card-sub">時安大廈(洋松街)業主立案法團</div>
            </div>
            <div class="work-card">
              <div class="work-card-title">目前大廈</div>
              <div class="work-stat-label">時安大廈</div>
              <div class="work-stat">4</div>
              <div class="work-stat-label">基本資料文件</div>
            </div>
            <div class="work-card">
              <div class="work-card-title">資料狀態</div>
              <div class="work-row">
                <div>
                  <strong>已有資料記錄</strong>
                  <span>可於下方同頁更新欄位與附件</span>
                </div>
                <span class="work-chip good">正常</span>
              </div>
            </div>
            <div class="work-card">
              <div class="work-card-title">大廈</div>
              <div class="work-card-sub">時安大廈</div>
            </div>
          </section>
          <section class="work-card">
            <div class="work-card-title">基本欄位</div>
            <div class="building-field-grid">
              <div
                v-for="f in buildingFields"
                :key="f.label"
                class="building-field"
              >
                <span>{{ f.label }}</span>
                <strong>{{ f.value }}</strong>
              </div>
            </div>
          </section>
          <section class="work-card">
            <div class="work-card-title">地圖網址</div>
            <iframe
              class="building-map-frame"
              src="https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d14763.34557254608!2d114.1557081846111!3d22.32202646614546!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x340400b7741874d5%3A0x92522a6b0453ab64!2z5aSn6KeS5ZKA5rSL5p2-6KGXNDbomZ_mmYLlronlpKflu4g!5e0!3m2!1szh-TW!2shk!4v1664891445389!5m2!1szh-TW!2shk"
              allowfullscreen
              loading="lazy"
              referrerpolicy="no-referrer-when-downgrade"
            />
          </section>
          <section class="work-card">
            <div class="work-card-title">基本文件區</div>
            <div class="building-doc-grid">
              <div
                v-for="d in buildingDocCards"
                :key="d.title"
                class="building-doc-card"
              >
                <div class="work-card-title">{{ d.title }}</div>
                <div class="work-card-sub">{{ d.desc }}</div>
                <div class="building-doc-count">{{ d.count }}</div>
                <div class="building-doc-empty">{{ d.empty }}</div>
              </div>
            </div>
          </section>
          <section class="work-card">
            <div class="building-section-head">
              <div class="work-card-title">表格</div>
              <button
                type="button"
                class="work-mini-btn primary"
              >
                新增
              </button>
            </div>
            <table class="work-table">
              <thead>
                <tr>
                  <th>標題</th>
                  <th>日期</th>
                  <th>月份</th>
                  <th>下載</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                <tr>
                  <td
                    class="building-empty-row"
                    colspan="5"
                  >
                    目前未有表格。
                  </td>
                </tr>
              </tbody>
            </table>
          </section>
          <section class="work-card">
            <div class="building-section-head">
              <div class="work-card-title">大廈資訊</div>
              <button
                type="button"
                class="work-mini-btn primary"
              >
                新增
              </button>
            </div>
            <table class="work-table">
              <thead>
                <tr>
                  <th>標題</th>
                  <th>日期</th>
                  <th>月份</th>
                  <th>下載</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                <tr>
                  <td
                    class="building-empty-row"
                    colspan="5"
                  >
                    目前未有大廈資訊。
                  </td>
                </tr>
              </tbody>
            </table>
          </section>
          <section class="work-card">
            <div class="building-section-head">
              <div class="work-card-title">平面圖</div>
              <button
                type="button"
                class="work-mini-btn primary"
              >
                新增
              </button>
            </div>
            <table class="work-table">
              <thead>
                <tr>
                  <th>標題</th>
                  <th>日期</th>
                  <th>月份</th>
                  <th>下載</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="p in floorPlans"
                  :key="p.title"
                >
                  <td>{{ p.title }}</td>
                  <td>{{ p.date }}</td>
                  <td>{{ p.month }}</td>
                  <td>
                    <a
                      class="building-link"
                      href="#"
                      @click.prevent="goBuildingDetail"
                    >查看</a>
                  </td>
                  <td>
                    <button
                      type="button"
                      class="work-mini-btn"
                    >
                      更新
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </section>
        </div>

        <!-- 大廈財務 -->
        <div
          v-show="activeTab === 'affairs-finance'"
          class="work-panel on"
        >
          <section class="work-hero">
            <div>
              <div class="work-kicker">Building Finance</div>
              <h2 class="work-title">大廈財務</h2>
              <p class="work-desc">按大廈查看管理費、財務報表、核數報告及繳款記錄。</p>
            </div>
          </section>
          <section class="work-card acct-card-wrap">
            <div class="acct-tabs">
              <button
                type="button"
                class="acct-tab"
                :class="{ on: financeSubTab === 'acct-overview' }"
                @click="switchFinanceSub('acct-overview')"
              >
                管理費總覽
              </button>
              <button
                type="button"
                class="acct-tab"
                :class="{ on: financeSubTab === 'acct-finance' }"
                @click="switchFinanceSub('acct-finance')"
              >
                財務報表
              </button>
              <button
                type="button"
                class="acct-tab"
                :class="{ on: financeSubTab === 'acct-audit' }"
                @click="switchFinanceSub('acct-audit')"
              >
                核數報告
              </button>
              <button
                type="button"
                class="acct-tab"
                :class="{ on: financeSubTab === 'acct-payment-records' }"
                @click="switchFinanceSub('acct-payment-records')"
              >
                管理費繳款記錄
              </button>
            </div>
            <div class="acct-body">
              <!-- 管理費總覽 -->
              <div
                v-show="financeSubTab === 'acct-overview'"
                class="acct-subpanel on"
              >
                <div class="acct-note">選用「電子付款」繳交管理費住戶，於3小時內即可結算，其他支付方式由於核對需時未能即時更新，敬請見諒。</div>
                <div class="acct-toolbar">
                  <div class="work-card-title acct-toolbar-title">管理費總覽</div>
                  <input
                    class="acct-search"
                    placeholder="Search.."
                  >
                </div>
                <div class="acct-table-wrap">
                  <table class="acct-table">
                    <thead>
                      <tr>
                        <th>單位</th>
                        <th>2025/11</th>
                        <th>2025/10</th>
                        <th>2025/09</th>
                        <th>2025/08月前</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr
                        v-for="r in financeOverview"
                        :key="r.unit"
                      >
                        <td>{{ r.unit }}</td>
                        <td :class="r.m11Status === 'due' ? 'acct-due' : 'acct-paid'">
                          {{ r.m11 }}
                        </td>
                        <td :class="r.m10Status === 'due' ? 'acct-due' : 'acct-paid'">
                          {{ r.m10 }}
                        </td>
                        <td :class="r.m09Status === 'due' ? 'acct-due' : 'acct-paid'">
                          {{ r.m09 }}
                        </td>
                        <td :class="r.mBeforeStatus === 'due' ? 'acct-due' : 'acct-paid'">
                          {{ r.mBefore }}
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
              <!-- 財務報表 -->
              <div
                v-show="financeSubTab === 'acct-finance'"
                class="acct-subpanel on"
              >
                <div class="acct-toolbar">
                  <div class="work-card-title acct-toolbar-title">法團財務報表</div>
                  <input
                    class="acct-search"
                    placeholder="Search.."
                  >
                </div>
                <div class="acct-table-wrap">
                  <table class="acct-table">
                    <thead>
                      <tr>
                        <th>日期</th>
                        <th>標題</th>
                        <th>查閱</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr
                        v-for="(r, i) in financeReports"
                        :key="i"
                      >
                        <td>{{ r.date }}</td>
                        <td>{{ r.title }}</td>
                        <td>
                          <button
                            type="button"
                            class="acct-download"
                          >
                            查閱
                          </button>
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
              <!-- 核數報告 -->
              <div
                v-show="financeSubTab === 'acct-audit'"
                class="acct-subpanel on"
              >
                <div class="acct-toolbar">
                  <div class="work-card-title acct-toolbar-title">核數報告</div>
                  <input
                    class="acct-search"
                    placeholder="Search.."
                  >
                </div>
                <div class="acct-table-wrap">
                  <table class="acct-table">
                    <thead>
                      <tr>
                        <th>日期</th>
                        <th>項目</th>
                        <th>查閱</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr
                        v-for="(r, i) in auditReports"
                        :key="i"
                      >
                        <td>{{ r.date }}</td>
                        <td>{{ r.item }}</td>
                        <td>
                          <button
                            type="button"
                            class="acct-download"
                          >
                            查閱
                          </button>
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
              <!-- 管理費繳款記錄 -->
              <div
                v-show="financeSubTab === 'acct-payment-records'"
                class="acct-subpanel on"
              >
                <div class="acct-dev">正在開發</div>
              </div>
            </div>
          </section>
        </div>

        <!-- 業戶帳目 -->
        <div
          v-show="activeTab === 'affairs-owner-account'"
          class="work-panel on"
        >
          <section class="work-hero">
            <div>
              <div class="work-kicker">Owner Account</div>
              <h2 class="work-title">業戶帳目</h2>
              <p class="work-desc">業戶帳目資料將於後續版本提供。</p>
            </div>
          </section>
          <div class="acct-dev">正在開發</div>
        </div>

        <!-- 申請表格 -->
        <div
          v-show="activeTab === 'affairs-forms'"
          class="work-panel on"
        >
          <section class="work-hero">
            <div>
              <div class="work-kicker">Forms</div>
              <h2 class="work-title">申請表格</h2>
              <p class="work-desc">選擇大廈後查看可下載的住戶及工程申請表格。</p>
            </div>
          </section>
          <section class="notice-admin-grid">
            <div class="notice-admin-card">
              <h3>機構</h3>
              <label class="notice-admin-label">選擇機構:</label>
              <select class="notice-select">
                <option>時安大廈(洋松街)業主立案法團</option>
              </select>
            </div>
            <div class="notice-admin-card">
              <h3>大廈</h3>
              <label class="notice-admin-label">選擇大廈:</label>
              <select
                v-model="selectedFormBuilding"
                class="notice-select"
              >
                <option
                  v-for="b in formBuildings"
                  :key="b"
                  :value="b"
                >
                  {{ b }}
                </option>
              </select>
            </div>
          </section>
          <section class="work-card">
            <div class="building-section-head">
              <div class="work-card-title">表格下載</div>
              <button
                type="button"
                class="work-mini-btn primary"
              >
                新增表格
              </button>
            </div>
            <table class="work-table">
              <thead>
                <tr>
                  <th>標題</th>
                  <th>查閱</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="f in forms"
                  :key="f.title"
                >
                  <td>{{ f.title }}</td>
                  <td>
                    <a
                      class="building-link"
                      href="#"
                      @click.prevent
                    >填寫表格</a>
                  </td>
                </tr>
              </tbody>
            </table>
          </section>
        </div>

        <!-- 意見提供 -->
        <div
          v-show="activeTab === 'affairs-feedback'"
          class="work-panel on"
        >
          <section class="work-hero">
            <div>
              <div class="work-kicker">Feedback</div>
              <h2 class="work-title">意見提供</h2>
              <p class="work-desc">選擇大廈、事項標題並填寫內容，提交後由管理處跟進。</p>
            </div>
          </section>
          <section class="work-card affairs-feedback-form">
            <div class="affairs-field">
              <label for="affairs-feedback-building">大廈</label>
              <select
                id="affairs-feedback-building"
                v-model="selectedFeedbackBuilding"
                class="affairs-select"
              >
                <option value="">---------</option>
                <option
                  v-for="b in feedbackBuildings"
                  :key="b"
                  :value="b"
                >
                  {{ b }}
                </option>
              </select>
            </div>
            <div class="affairs-field">
              <label for="affairs-feedback-title">標題</label>
              <select
                id="affairs-feedback-title"
                v-model="selectedFeedbackTitle"
                class="affairs-select"
              >
                <option value="">---------</option>
                <option
                  v-for="t in feedbackTitles"
                  :key="t"
                  :value="t"
                >
                  {{ t }}
                </option>
              </select>
            </div>
            <div class="affairs-field">
              <label for="affairs-feedback-content">內容</label>
              <textarea
                id="affairs-feedback-content"
                v-model="feedbackContent"
                class="affairs-textarea"
              />
            </div>
            <div class="affairs-submit-row">
              <button
                type="button"
                class="work-action"
                @click="submitFeedback"
              >
                提交
              </button>
              <span class="work-card-sub">閣下提供之意見將絕對保密。</span>
            </div>
          </section>
          <section class="work-card">
            <div class="work-card-title">最近記錄</div>
            <table class="work-table">
              <thead>
                <tr>
                  <th>事項</th>
                  <th>分類</th>
                  <th>狀態</th>
                  <th>更新時間</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="(r, i) in feedbackRecords"
                  :key="i"
                >
                  <td>{{ r.subject }}</td>
                  <td>{{ r.category }}</td>
                  <td>
                    <span
                      class="work-chip"
                      :class="r.status === 'processing' ? 'warn' : 'good'"
                    >{{ r.statusText }}</span>
                  </td>
                  <td>{{ r.updatedAt }}</td>
                </tr>
              </tbody>
            </table>
          </section>
        </div>

        <!-- 智能門禁 -->
        <div
          v-show="activeTab === 'affairs-access'"
          class="work-panel on"
        >
          <section class="work-hero">
            <div>
              <div class="work-kicker">Access</div>
              <h2 class="work-title">智能門禁</h2>
              <p class="work-desc">此功能暫未開通。</p>
            </div>
          </section>
          <section class="work-card">
            <div class="work-card-title">暫未開通</div>
            <div class="work-card-sub">智能門禁、訪客通行與門鎖聯動功能目前尚未開放。</div>
          </section>
        </div>

        <!-- 視像監控 -->
        <div
          v-show="activeTab === 'affairs-icctv'"
          class="work-panel on"
        >
          <section class="work-hero">
            <div>
              <div class="work-kicker">ICCTV</div>
              <h2 class="work-title">視像監控</h2>
              <p class="work-desc">選擇大廈與鏡頭後，查看即時監控畫面。</p>
            </div>
            <button
              type="button"
              class="work-action"
            >
              新窗口
            </button>
          </section>
          <section class="work-card icctv-panel-wide">
            <div class="icctv-building-select">
              <select class="staff-select">
                <option>請選擇大廈</option>
              </select>
              <button
                type="button"
                class="work-mini-btn"
              >
                重新載入
              </button>
            </div>
            <div class="icctv-layout">
              <div>
                <div class="icctv-building-meta">尚未選擇大廈</div>
                <div class="icctv-camera-list">
                  <div class="icctv-empty-camera">尚未提供鏡頭</div>
                </div>
                <p class="icctv-note">來源頁如限制嵌入，可使用新窗口開啟。</p>
              </div>
              <div>
                <div class="icctv-viewer-bar">
                  <div class="icctv-viewer-meta">尚未選擇鏡頭</div>
                  <div class="icctv-viewer-actions">
                    <button
                      type="button"
                      class="work-mini-btn"
                    >
                      新窗口
                    </button>
                  </div>
                </div>
                <div class="icctv-viewer">
                  <div class="icctv-empty">
                    <strong>未選擇鏡頭</strong>
                    <span>選擇大廈與鏡頭後，監控畫面會顯示在此處。</span>
                  </div>
                </div>
              </div>
            </div>
          </section>
        </div>

        <!-- 設備監測 -->
        <div
          v-show="activeTab === 'affairs-equipment'"
          class="work-panel on"
        >
          <section class="work-hero">
            <div>
              <div class="work-kicker">Equipment</div>
              <h2 class="work-title">設備監測</h2>
              <p class="work-desc">查看大廈設備狀態、巡檢紀錄與異常提醒。</p>
            </div>
            <button
              type="button"
              class="work-action secondary"
            >
              重新整理
            </button>
          </section>
          <section class="work-card">
            <div class="work-card-title">設備狀態</div>
            <table class="work-table">
              <thead>
                <tr>
                  <th>設備</th>
                  <th>位置</th>
                  <th>狀態</th>
                  <th>更新時間</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="(r, i) in equipmentRows"
                  :key="i"
                >
                  <td>{{ r.device }}</td>
                  <td>{{ r.location }}</td>
                  <td>
                    <span
                      class="work-chip"
                      :class="r.status"
                    >{{ r.statusText }}</span>
                  </td>
                  <td>{{ r.updatedAt }}</td>
                </tr>
              </tbody>
            </table>
          </section>
        </div>
      </main>
    </div>
  </div>
</template>

<style scoped>
/* Layout shell */
.page-affairs {
  min-height: calc(100vh - var(--nav-h, 52px));
  background: rgb(var(--color-surface-2));
}

.work-shell {
  display: grid;
  grid-template-columns: 240px minmax(0, 1fr);
  gap: 16px;
  max-width: 1180px;
  margin: 0 auto;
  padding: 12px 16px 16px;
  color: rgb(var(--color-text));
}

/* Sidebar */
.work-sidebar {
  position: sticky;
  top: calc(var(--nav-h, 52px) + 12px);
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

.work-nav {
  display: grid;
  gap: 2px;
  margin-top: 18px;
}

.work-nav-item {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: space-between;
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
  transition: background-color 0.15s ease, color 0.15s ease;
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

/* Main area */
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

/* Hero */
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

.work-action.secondary {
  border: 1px solid rgb(var(--color-border));
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
}

/* Card */
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

.work-card-sub {
  color: rgb(var(--color-ink-3));
  font-size: 12px;
  line-height: 1.7;
}

.work-stat {
  color: rgb(var(--color-primary));
  font-size: 24px;
  font-weight: 700;
  line-height: 1.1;
}

.work-stat-label {
  margin-top: 6px;
  color: rgb(var(--color-ink-3));
  font-size: 12px;
}

.work-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  border-top: 1px solid rgb(var(--color-surface-3));
  padding: 12px 0;
}

.work-row:first-child {
  border-top: 0;
  padding-top: 0;
}

.work-row:last-child {
  padding-bottom: 0;
}

.work-row strong {
  display: block;
  color: rgb(var(--color-text));
  font-size: 13px;
  font-weight: 600;
}

.work-row span {
  display: block;
  margin-top: 3px;
  color: rgb(var(--color-ink-3));
  font-size: 12px;
  line-height: 1.5;
}

/* Chip */
.work-chip {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  background: rgb(var(--color-surface-2));
  color: rgb(var(--color-ink-2));
  font-size: 11px;
  font-weight: 600;
  padding: 5px 9px;
  white-space: nowrap;
}

.work-chip.good {
  background: rgb(var(--color-success-bg));
  color: rgb(var(--color-success));
}

.work-chip.warn {
  background: rgb(var(--color-warning-bg));
  color: rgb(var(--color-warning));
}

.work-chip.brand {
  background: rgb(var(--color-primary-soft));
  color: rgb(var(--color-primary));
}

/* Table */
.work-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
}

.work-table th {
  border-bottom: 1px solid rgb(var(--color-border));
  color: rgb(var(--color-ink-3));
  font-weight: 500;
  padding: 10px;
  text-align: left;
}

.work-table td {
  border-bottom: 1px solid rgb(var(--color-surface-3));
  padding: 12px 10px;
  color: rgb(var(--color-ink-2));
}

.work-table tr:last-child td {
  border-bottom: 0;
}

.work-table-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

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

.work-mini-btn.primary {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: #fff;
}

/* Notice admin */
.notice-admin-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 12px;
}

.notice-admin-card {
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 14px 16px;
}

.notice-admin-card h3 {
  margin: 0 0 8px;
  color: rgb(var(--color-text));
  font-size: 13px;
  font-weight: 600;
}

.notice-admin-label {
  display: block;
  margin-bottom: 6px;
  color: rgb(var(--color-ink-3));
  font-size: 12px;
}

.notice-select {
  width: 100%;
  height: 34px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 6px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
  font-family: inherit;
  font-size: 12px;
  padding: 0 10px;
}

.notice-current {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
  color: rgb(var(--color-ink-3));
  font-size: 12px;
}

.notice-action-btn {
  border: 1px solid rgb(var(--color-border));
  border-radius: 6px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-ink-2));
  cursor: pointer;
  font-family: inherit;
  font-size: 12px;
  font-weight: 600;
  padding: 6px 10px;
}

.notice-action-btn:hover {
  border-color: rgb(var(--color-brand-mid));
  color: rgb(var(--color-primary));
}

.notice-table {
  table-layout: auto;
}

/* Building profile */
.building-summary-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 12px;
}

.building-field-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 12px;
}

.building-field {
  display: grid;
  gap: 4px;
  border-top: 1px solid rgb(var(--color-surface-3));
  padding: 10px 0;
}

.building-field span {
  color: rgb(var(--color-ink-3));
  font-size: 12px;
}

.building-field strong {
  color: rgb(var(--color-text));
  font-size: 13px;
  font-weight: 600;
  word-break: break-word;
}

.building-map-frame {
  width: 100%;
  height: 280px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 6px;
  background: rgb(var(--color-surface-2));
}

.building-doc-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 12px;
}

.building-doc-card {
  border: 1px solid rgb(var(--color-border));
  border-radius: 6px;
  background: rgb(var(--color-surface-2));
  padding: 14px;
}

.building-doc-card .work-card-title {
  margin-bottom: 6px;
}

.building-doc-count {
  margin-top: 10px;
  color: rgb(var(--color-primary));
  font-size: 13px;
  font-weight: 700;
}

.building-doc-empty {
  margin-top: 6px;
  color: rgb(var(--color-ink-3));
  font-size: 12px;
  line-height: 1.6;
}

.building-section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 10px;
}

.building-section-head .work-card-title {
  margin: 0;
}

.building-empty-row {
  text-align: center;
  color: rgb(var(--color-ink-3));
  padding: 24px 10px;
}

.building-link {
  color: rgb(var(--color-primary));
  font-size: 12px;
  font-weight: 600;
  text-decoration: none;
}

.building-link:hover {
  text-decoration: underline;
}

/* Accounting */
.acct-card-wrap {
  padding: 0;
  overflow: hidden;
}

.acct-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 0;
  border-bottom: 1px solid rgb(var(--color-border));
  background: rgb(var(--color-surface-2));
}

.acct-tab {
  border: 0;
  border-bottom: 2px solid transparent;
  background: transparent;
  color: rgb(var(--color-ink-3));
  cursor: pointer;
  font-family: inherit;
  font-size: 13px;
  font-weight: 500;
  padding: 12px 16px;
  transition: color 0.15s ease, border-color 0.15s ease;
}

.acct-tab:hover {
  color: rgb(var(--color-text));
}

.acct-tab.on {
  color: rgb(var(--color-primary));
  border-bottom-color: rgb(var(--color-primary));
  font-weight: 600;
}

.acct-body {
  padding: 16px;
}

.acct-subpanel.on {
  display: grid;
  gap: 12px;
}

.acct-note {
  border: 1px solid rgb(var(--color-warning-bg));
  border-radius: 6px;
  background: rgb(var(--color-warning-bg));
  color: rgb(var(--color-warning));
  font-size: 12px;
  line-height: 1.6;
  padding: 10px 12px;
}

.acct-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.acct-toolbar-title {
  margin: 0;
}

.acct-search {
  width: 220px;
  max-width: 40%;
  height: 32px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 6px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
  font-family: inherit;
  font-size: 12px;
  padding: 0 10px;
}

.acct-table-wrap {
  overflow: auto;
  border: 1px solid rgb(var(--color-border));
  border-radius: 6px;
  background: rgb(var(--color-surface));
}

.acct-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
}

.acct-table th {
  border-bottom: 1px solid rgb(var(--color-border));
  background: rgb(var(--color-surface-2));
  color: rgb(var(--color-ink-3));
  font-weight: 600;
  padding: 10px;
  text-align: left;
  white-space: nowrap;
}

.acct-table td {
  border-bottom: 1px solid rgb(var(--color-surface-3));
  color: rgb(var(--color-ink-2));
  padding: 10px;
  white-space: nowrap;
}

.acct-table tr:last-child td {
  border-bottom: 0;
}

.acct-paid {
  color: rgb(var(--color-success));
  font-weight: 600;
}

.acct-due {
  color: rgb(var(--color-danger));
  font-weight: 600;
}

.acct-download {
  border: 1px solid rgb(var(--color-border));
  border-radius: 6px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-ink-2));
  cursor: pointer;
  font-family: inherit;
  font-size: 12px;
  font-weight: 600;
  padding: 5px 10px;
}

.acct-dev {
  border: 1px dashed rgb(var(--color-border-2));
  border-radius: 8px;
  background: rgb(var(--color-surface-2));
  color: rgb(var(--color-ink-3));
  font-size: 13px;
  text-align: center;
  padding: 32px 16px;
}

/* Feedback form */
.affairs-feedback-form {
  display: grid;
  gap: 14px;
}

.affairs-field {
  display: grid;
  gap: 6px;
}

.affairs-field label {
  color: rgb(var(--color-ink-3));
  font-size: 12px;
}

.affairs-select {
  width: 100%;
  height: 36px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 6px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
  font-family: inherit;
  font-size: 13px;
  padding: 0 10px;
}

.affairs-textarea {
  width: 100%;
  min-height: 120px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 6px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
  font-family: inherit;
  font-size: 13px;
  line-height: 1.6;
  padding: 10px 12px;
  resize: vertical;
}

.affairs-submit-row {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

/* ICCTV */
.icctv-panel-wide {
  display: grid;
  gap: 12px;
}

.icctv-building-select {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.staff-select {
  min-width: 200px;
  height: 34px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 6px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
  font-family: inherit;
  font-size: 12px;
  padding: 0 10px;
}

.icctv-layout {
  display: grid;
  grid-template-columns: minmax(0, 280px) minmax(0, 1fr);
  gap: 16px;
}

.icctv-building-meta {
  border: 1px solid rgb(var(--color-border));
  border-radius: 6px;
  background: rgb(var(--color-surface-2));
  color: rgb(var(--color-ink-3));
  font-size: 12px;
  padding: 10px 12px;
}

.icctv-camera-list {
  margin-top: 8px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 6px;
  background: rgb(var(--color-surface));
  min-height: 120px;
  padding: 10px;
}

.icctv-empty-camera {
  color: rgb(var(--color-ink-3));
  font-size: 12px;
  text-align: center;
  padding: 32px 0;
}

.icctv-note {
  margin: 8px 0 0;
  color: rgb(var(--color-ink-3));
  font-size: 11px;
  line-height: 1.6;
}

.icctv-viewer-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  border: 1px solid rgb(var(--color-border));
  border-bottom: 0;
  border-radius: 6px 6px 0 0;
  background: rgb(var(--color-surface-2));
  padding: 8px 12px;
}

.icctv-viewer-meta {
  color: rgb(var(--color-ink-3));
  font-size: 12px;
}

.icctv-viewer-actions {
  display: flex;
  gap: 8px;
}

.icctv-viewer {
  position: relative;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0 0 6px 6px;
  background: rgb(var(--color-surface-3));
  min-height: 260px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.icctv-empty {
  display: grid;
  gap: 6px;
  text-align: center;
  color: rgb(var(--color-ink-3));
  font-size: 12px;
}

.icctv-empty strong {
  color: rgb(var(--color-ink-2));
  font-size: 14px;
  font-weight: 600;
}

/* Responsive */
@media (max-width: 1023px) {
  .work-shell {
    grid-template-columns: 1fr;
    padding: 14px;
  }

  .work-sidebar {
    position: static;
  }

  .work-nav {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .icctv-layout {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 767px) {
  .work-shell {
    padding: 12px;
    gap: 12px;
  }

  .work-nav {
    grid-template-columns: 1fr;
  }

  .work-hero {
    align-items: flex-start;
    flex-direction: column;
  }

  .work-title {
    font-size: 18px;
  }

  .building-summary-grid,
  .building-field-grid,
  .building-doc-grid,
  .notice-admin-grid {
    grid-template-columns: 1fr;
  }

  .acct-search {
    width: 100%;
    max-width: none;
  }

  .acct-toolbar {
    flex-direction: column;
    align-items: stretch;
  }

  .affairs-submit-row {
    flex-direction: column;
    align-items: stretch;
  }

  .affairs-submit-row .work-action {
    width: 100%;
  }
}
</style>
