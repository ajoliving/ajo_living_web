<!--
 * 會員中心頁。
 * 1. 承載左側模組導覽與右側內容面板，支援帳號管理、授權副戶、物業綁定、錢包、聊天、樓盤、住宅、家具、收藏、訂單等面板切換。
 * 2. 所有資料為靜態 mock，不呼叫任何 API。
 * 3. 響應式設計：桌面雙欄、行動單欄。
-->
<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';

const router = useRouter();

// 1. 面板索引類型
type PanelKey =
  | 'profile-account'
  | 'profile-subaccounts'
  | 'profile-property-binding'
  | 'profile-wallet'
  | 'profile-chat'
  | 'profile-properties'
  | 'profile-homes'
  | 'profile-furniture'
  | 'profile-saved'
  | 'profile-orders';

// 2. 導覽項目
const navItems: { key: PanelKey; label: string }[] = [
  { key: 'profile-account', label: '帳號管理' },
  { key: 'profile-subaccounts', label: '授權副戶' },
  { key: 'profile-property-binding', label: '物業綁定' },
  { key: 'profile-wallet', label: 'AJO 錢包' },
  { key: 'profile-chat', label: '聊天' },
  { key: 'profile-properties', label: '我的樓盤' },
  { key: 'profile-homes', label: '我的住宅' },
  { key: 'profile-furniture', label: '我的家具' },
  { key: 'profile-saved', label: '我的收藏' },
  { key: 'profile-orders', label: '交易訂單' },
];

// 3. 當前面板
const activePanel = ref<PanelKey>('profile-account');

// 4. 切換面板
const switchPanel = (key: PanelKey) => {
  activePanel.value = key;
};

// 5. 帳號管理 mock 資料
const accountProfile = {
  displayName: 'patrick',
  ismartAccount: 'patrick',
  email: '尚未綁定電郵',
  phone: 'ismart 10',
  password: '未設定',
  publisherRole: '未設定',
  regionCode: 'unknown',
  managedBuildings:
    '康睦庭園第二座, 萬高大廈B座, 僑偉大廈, 利來大廈, 仁美大廈, 豐富大廈, 鴻英大廈, 東昇樓, 華園, 麗麗大廈, 協和大廈, 景暉閣, 新萬利大廈, 東山臺24號, 仁英大廈, 浣紗大廈, 東南大樓(77號), 東南大樓(75號), 華興大廈(269號), 華興大廈(271號), 得運大廈, 仁文大廈, 富澤軒, 天富大廈, 榮森工業第二大廈, 仁利大廈, 嘉樂苑, 南昌苑, 測試2大廈, 測試1大廈, 坪麗苑, 玉桂園(車位), 玉桂園(1座), 玉桂園(2座), 玉桂園(3座), 玉桂園(4座), 玉桂園(5座), 玉桂園(6座), 玉桂園(7座), 玉桂園(8座), 玉桂園(9座), 玉桂園(10座), 玉桂園(11座)',
};

const identityStatus = {
  memberNo: '01KRWKDTQY6DY1H650GB63AC66',
  memberStatus: 'active',
  memberType: 'user',
  primaryRole: 'staff',
  profileCompleteness: '已完成',
  staffPermission: '已啟用',
};

const bindCurrent = '康睦庭園第二座 / 02 / D';

const ismartAccountData = [
  { label: '帳戶編號', value: 'SAWYER' },
  { label: '帳戶電話', value: '90771352' },
  { label: '帳戶電郵', value: 'sawyer@seventy2.hk' },
  { label: '業戶名稱(英)', value: 'CHEUNG CHUN HO' },
  { label: '業戶名稱(中)', value: '未設定' },
  { label: '證件編號', value: '********' },
];

const ismartHouseholdData = [
  { label: '法體類型', value: '自然人' },
  { label: '性別', value: 'M' },
  { label: '出生日期', value: '未設定' },
  { label: '聯絡人名稱', value: 'S' },
  { label: '聯絡人電話', value: '90771352' },
  { label: '帳單電郵', value: 'sawyer@seventy2.hk' },
  { label: '帳單地址', value: 'test@123' },
];

const relatedProperties = [
  { name: '界限大廈', status: '登記業主' },
  { name: '協和大廈', status: '法團' },
  { name: '華興大廈(271號)', status: '登記業主' },
  { name: '華興大廈(269號)', status: '登記業主' },
  { name: '仁英大廈 G 02', status: '登記業主' },
];

// 6. 授權副戶 mock 資料
const subaccountGroups = [
  {
    name: '華興大廈(271號)',
    rows: [
      { location: '華興大廈(271號)', user: 'The Incorporated Owners of No. 269, 271 Temple Street', type: '法團', permission: '只讀' },
    ],
  },
  {
    name: '仁英大廈 G 02',
    rows: [
      { location: '仁英大廈 G 02', user: '住客帳戶', type: '住客', permission: '只讀' },
    ],
  },
  { name: '仁英大廈 07 B', rows: [] },
  { name: '仁英大廈 08 C', rows: [] },
];

// 7. 物業綁定 mock 資料
const bindingSteps = [
  { num: '1', title: '選擇平台', desc: '先選擇 AJO PM 大廈平台或 AJO Rent 租務平台。' },
  { num: '2', title: '填寫聯絡資料', desc: '提供申請人姓名、電話及電郵，便於管理處核對。' },
  { num: '3', title: '上載證明文件', desc: '按身份提交業權、租約或授權文件。' },
  { num: '4', title: '等待審批', desc: '審批通過後，該物業會加入帳戶可見範圍。' },
];

const bindingPlatform = ref<'pm' | 'rent'>('pm');

const bindingDocs = [
  { title: '身份證明', desc: '身份證、護照或公司授權人身份文件。' },
  { title: '物業關係證明', desc: '業權文件、租約、住戶證明或授權書。' },
  { title: '最近賬單或收據', desc: '管理費賬單、水電煤賬單或管理處認可文件。' },
  { title: '補充文件', desc: '如管理處要求，可上載其他補充證明。' },
];

const bindingStatusList = [
  { title: '康睦庭園第二座 / 02 / D', meta: '業主身份 · 已於 2026年6月5日完成審批', chip: '已綁定', chipType: 'good' },
  { title: '仁英大廈 / 07 / B', meta: '租客身份 · 等待管理處核對文件', chip: '審批中', chipType: 'warn' },
];

// 8. 錢包 mock 資料
const walletStats = { earned: '11,050', used: '2,000' };

const walletTransactions = [
  { type: '增加', source: 'wallet · ad_reward', points: '50 AJO Point', date: '2026年6月5日' },
  { type: '扣除', source: 'property_sale · publish', points: '1,000 AJO Point', date: '2026年5月22日' },
  { type: '增加', source: 'wallet · operator_grant', points: '10,000 AJO Point', date: '2026年5月22日' },
];

// 9. 聊天 mock 資料
const chatThreads = [
  { id: 't1', avatar: '租', name: '租務查詢', time: '12:30', preview: '請確認預約睇樓時間。', unread: true, active: true },
  { id: 't2', avatar: '家', name: '家具買家', time: '昨天', preview: '已收到交收地點資料。', unread: false, active: false },
  { id: 't3', avatar: '管', name: '大廈管理處', time: '6月5日', preview: '管理費收據已更新。', unread: false, active: false },
];

const chatMessages = [
  { day: '今天', rows: [
    { text: '你好，想預約今晚 7:30 睇樓，請問時間是否可以確認？', time: '12:18', sent: false },
    { text: '可以，已為你保留今晚 7:30，稍後會發送到達資料。', time: '12:24', sent: true },
    { text: '收到，請確認集合位置。', time: '12:30', sent: false },
  ] },
];

// 10. 樓盤 mock 資料
const propertyListings = [
  { title: '佐敦 高臨 代理盤', area: '高臨', price: 'HK$0', status: '上架中', chipType: 'good' },
  { title: '佐敦道38號 唐三樓 套房', area: 'Kowloon', price: 'HK$0', status: '草稿', chipType: '' },
];

// 11. 住宅 mock 資料
const homeListings = [
  { name: '康睦庭園第二座 / 02 / D', identity: 'staff', status: '已綁定', chipType: 'good', updatedAt: '2026年6月5日' },
  { name: 'Harbour Residence', identity: '申請人', status: '待確認', chipType: 'warn', updatedAt: '2026年5月22日' },
];

// 12. 家具 mock 資料
const furnitureListings = [
  { name: '北歐實木餐桌', category: '家居傢俱', price: 'HK$2,400', status: '公開', chipType: 'good' },
  { name: 'LG 洗衣機 8kg', category: '家庭電器', price: 'HK$1,800', status: '草稿', chipType: '' },
];

// 13. 收藏 mock 資料
const savedItems = [
  { name: '佐敦高級住宅', category: '樓盤', price: 'HK$36,000/月', savedAt: '今天 12:30' },
  { name: '纖柔牙刷 精巧頭 3支裝', category: '綜合優惠', price: 'HK$14.50', savedAt: '昨天 17:20' },
];

// 14. 訂單 mock 資料
const orderListings = [
  { id: 'AJO-20260605-01', type: '管理費', amount: 'HK$2,850', status: '待繳', chipType: 'warn' },
  { id: 'AJO-20260522-03', type: '樓盤發布', amount: 'HK$1,000', status: '已完成', chipType: 'good' },
];

// 15. 退出登入
const handleLogout = () => {
  router.push('/');
};
</script>

<template>
  <main class="account-my-page">
    <div class="work-shell">
      <!-- 1. 左側導覽 -->
      <aside class="work-sidebar">
        <h1>會員中心</h1>
        <nav class="work-nav">
          <button
            v-for="item in navItems"
            :key="item.key"
            type="button"
            class="work-nav-item"
            :class="{ on: activePanel === item.key }"
            @click="switchPanel(item.key)"
          >
            {{ item.label }}
          </button>
        </nav>
      </aside>

      <!-- 2. 右側內容 -->
      <div class="work-main">
        <!-- 2.1 帳號管理 -->
        <div v-show="activePanel === 'profile-account'" class="work-panel on">
          <section class="work-hero">
            <div class="work-account-topbar">
              <div class="work-account-head">
                <div class="work-account-user">
                  <div class="work-account-avatar">P</div>
                  <div>
                    <div class="work-account-name">patrick</div>
                    <div class="work-account-sub">個人資料</div>
                  </div>
                </div>
              </div>
              <div class="work-account-actions">
                <button type="button" class="work-account-btn outline" @click="handleLogout">退出登入</button>
                <button type="button" class="work-account-btn primary">編輯資料</button>
                <button type="button" class="work-account-btn">更新 iSmart</button>
              </div>
            </div>
          </section>

          <section class="work-account-grid">
            <div class="work-account-card">
              <div class="work-card-title">帳戶資料</div>
              <div class="work-account-field"><span>顯示名稱</span><strong>{{ accountProfile.displayName }}</strong></div>
              <div class="work-account-field"><span>ISMART 帳戶</span><strong>{{ accountProfile.ismartAccount }}</strong></div>
              <div class="work-account-field"><span>電郵地址</span><strong>{{ accountProfile.email }}</strong></div>
              <div class="work-account-field"><span>電話號碼</span><strong>{{ accountProfile.phone }}</strong></div>
              <div class="work-account-field"><span>登入密碼</span><strong>{{ accountProfile.password }}</strong></div>
              <div class="work-account-field"><span>發布者身份</span><strong>{{ accountProfile.publisherRole }}</strong></div>
              <div class="work-account-field"><span>地區代碼</span><strong>{{ accountProfile.regionCode }}</strong></div>
              <div class="work-account-field"><span>員工管理屋苑</span><strong>{{ accountProfile.managedBuildings }}</strong></div>
            </div>
            <div class="work-account-card">
              <div class="work-card-title">身份狀態</div>
              <div class="work-account-field"><span>會員編號</span><strong>{{ identityStatus.memberNo }}</strong></div>
              <div class="work-account-field"><span>會員狀態</span><strong>{{ identityStatus.memberStatus }}</strong></div>
              <div class="work-account-field"><span>會員類型</span><strong>{{ identityStatus.memberType }}</strong></div>
              <div class="work-account-field"><span>主要角色</span><strong>{{ identityStatus.primaryRole }}</strong></div>
              <div class="work-account-field"><span>資料完整度</span><strong>{{ identityStatus.profileCompleteness }}</strong></div>
              <div class="work-account-field"><span>員工權限</span><strong>{{ identityStatus.staffPermission }}</strong></div>
              <div class="work-role-strip">
                <span class="work-role-badge">staff</span>
                <span class="work-role-tag">尚未分配權限</span>
              </div>
            </div>
          </section>

          <section class="work-card work-card--spaced">
            <div class="work-card-title">綁定單位</div>
            <div class="work-bind-wrap">
              <div class="work-bind-current">{{ bindCurrent }}</div>
              <div class="work-bind-grid">
                <div class="work-bind-field">
                  <label class="work-bind-label">大廈</label>
                  <select class="work-bind-select">
                    <option>康睦庭園第二座</option>
                    <option>協和大廈</option>
                    <option>仁英大廈</option>
                  </select>
                </div>
                <div class="work-bind-field">
                  <label class="work-bind-label">樓層</label>
                  <select class="work-bind-select">
                    <option>02</option>
                    <option>03</option>
                    <option>05</option>
                  </select>
                </div>
                <div class="work-bind-field">
                  <label class="work-bind-label">單位</label>
                  <select class="work-bind-select">
                    <option>D</option>
                    <option>A</option>
                    <option>B</option>
                  </select>
                </div>
              </div>
              <div class="work-bind-actions">
                <button type="button" class="work-action work-compact-action">儲存單位</button>
              </div>
            </div>
          </section>

          <section class="work-card work-card--spaced">
            <div class="work-card-title">iSmart 帳號資料</div>
            <div class="work-ismart-grid">
              <div class="work-ismart-card">
                <div class="work-card-sub">帳號資料</div>
                <div v-for="item in ismartAccountData" :key="item.label" class="work-row">
                  <div>
                    <strong>{{ item.label }}</strong>
                    <span>{{ item.value }}</span>
                  </div>
                </div>
              </div>
              <div class="work-ismart-card">
                <div class="work-card-sub">業戶資料</div>
                <div v-for="item in ismartHouseholdData" :key="item.label" class="work-row">
                  <div>
                    <strong>{{ item.label }}</strong>
                    <span>{{ item.value }}</span>
                  </div>
                </div>
              </div>
            </div>
          </section>

          <section class="work-card work-card--spaced">
            <div class="work-card-title">相關物業</div>
            <div class="work-table-wrap">
              <table class="work-table">
                <thead><tr><th>物業</th><th>狀態</th></tr></thead>
                <tbody>
                  <tr v-for="item in relatedProperties" :key="item.name">
                    <td>{{ item.name }}</td>
                    <td>{{ item.status }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </section>

          <section class="work-card work-card--spaced">
            <div class="work-card-title">提示設定</div>
            <div class="work-setting-box">
              <label class="work-setting-check">
                <input type="checkbox" checked>
                <span>接收大廈通告電郵提示</span>
              </label>
              <button type="button" class="work-action work-compact-action">提交</button>
            </div>
          </section>
        </div>

        <!-- 2.2 授權副戶 -->
        <div v-show="activePanel === 'profile-subaccounts'" class="work-panel on">
          <section class="work-hero">
            <div>
              <div class="work-kicker">Subaccounts</div>
              <h2 class="work-title">授權副戶</h2>
              <p class="work-desc">只讀顯示各物業的授權副戶資料，後續再接同步與管理操作。</p>
            </div>
          </section>
          <section class="work-subaccount-group">
            <div v-for="group in subaccountGroups" :key="group.name" class="work-subaccount-card">
              <div class="work-subaccount-head">
                <div class="work-subaccount-name">{{ group.name }}</div>
                <button type="button" class="work-action">新增授權</button>
              </div>
              <table class="work-table">
                <thead><tr><th>地點</th><th>用戶</th><th>類型</th><th>權限</th></tr></thead>
                <tbody>
                  <tr v-for="row in group.rows" :key="row.location">
                    <td>{{ row.location }}</td>
                    <td>{{ row.user }}</td>
                    <td>{{ row.type }}</td>
                    <td>{{ row.permission }}</td>
                  </tr>
                  <tr v-if="group.rows.length === 0">
                    <td colspan="4" class="work-subaccount-empty">目前未有授權副戶資料</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </section>
        </div>

        <!-- 2.3 物業綁定 -->
        <div v-show="activePanel === 'profile-property-binding'" class="work-panel on">
          <section class="work-hero">
            <div>
              <div class="work-kicker">Property Binding</div>
              <h2 class="work-title">物業綁定</h2>
              <p class="work-desc">提交大廈與單位資料，並上載指定文件予管理處審批。</p>
            </div>
          </section>
          <section class="work-card">
            <div class="work-card-title">申請流程</div>
            <div class="binding-flow">
              <div v-for="step in bindingSteps" :key="step.num" class="binding-step">
                <div class="binding-step-num">{{ step.num }}</div>
                <div class="binding-step-title">{{ step.title }}</div>
                <div class="binding-step-desc">{{ step.desc }}</div>
              </div>
            </div>
          </section>
          <section class="work-card">
            <div class="work-card-title">選擇平台</div>
            <div class="binding-platform-grid">
              <button
                type="button"
                class="binding-platform-card"
                :class="{ on: bindingPlatform === 'pm' }"
                @click="bindingPlatform = 'pm'"
              >
                <strong>AJO PM 大廈平台</strong>
                <span>適用於已接入 iSmart 或由物業管理公司審批的大廈。</span>
              </button>
              <button
                type="button"
                class="binding-platform-card"
                :class="{ on: bindingPlatform === 'rent' }"
                @click="bindingPlatform = 'rent'"
              >
                <strong>AJO Rent 租務平台</strong>
                <span>適用於租務住宅或由 AJO 確認的租住申請。</span>
              </button>
            </div>
            <div class="binding-review-box">
              <strong>審批責任</strong>
              <span v-if="bindingPlatform === 'pm'">AJO PM 大廈平台由物業管理公司審批。</span>
              <span v-else>AJO Rent 租務平台由 AJO 根據租務資料確認。</span>
            </div>
          </section>
          <section class="work-card">
            <div class="work-card-title">綁定申請</div>
            <div v-show="bindingPlatform === 'pm'" class="binding-platform-panel on">
              <div class="binding-subtitle">已支援大廈</div>
              <div class="acct-form-grid">
                <div class="acct-field full">
                  <label>大廈</label>
                  <select class="acct-select">
                    <option>請選擇已支援大廈</option>
                    <option>時安大廈</option>
                    <option>仁英大廈</option>
                    <option>康睦庭園第二座</option>
                    <option>協和大廈</option>
                  </select>
                </div>
              </div>
              <div class="binding-subtitle">未收錄大廈資料</div>
              <div class="acct-form-grid">
                <div class="acct-field"><label>大廈名稱</label><input class="acct-input" type="text" placeholder="請輸入大廈名稱"></div>
                <div class="acct-field"><label>大廈地址</label><input class="acct-input" type="text" placeholder="請輸入完整地址"></div>
                <div class="acct-field"><label>樓層</label><input class="acct-input" type="text" placeholder="例如 07"></div>
                <div class="acct-field"><label>單位</label><input class="acct-input" type="text" placeholder="例如 B"></div>
                <div class="acct-field"><label>申請身份</label><select class="acct-select"><option>業主</option><option>租客</option><option>住戶代表</option><option>公司授權人</option></select></div>
              </div>
            </div>
            <div v-show="bindingPlatform === 'rent'" class="binding-platform-panel on">
              <div class="binding-subtitle">租務單位資料</div>
              <div class="acct-form-grid">
                <div class="acct-field"><label>租務大廈或項目</label><input class="acct-input" type="text" placeholder="請輸入大廈或項目名稱"></div>
                <div class="acct-field"><label>申請身份</label><select class="acct-select"><option>租客</option><option>住戶代表</option><option>公司授權人</option></select></div>
                <div class="acct-field"><label>樓層</label><input class="acct-input" type="text" placeholder="例如 07"></div>
                <div class="acct-field"><label>單位</label><input class="acct-input" type="text" placeholder="例如 B"></div>
              </div>
            </div>
            <div class="acct-form-grid binding-common-grid">
              <div class="acct-field"><label>申請人姓名</label><input class="acct-input" type="text"></div>
              <div class="acct-field"><label>聯絡電話</label><input class="acct-input" type="text"></div>
              <div class="acct-field full"><label>聯絡電郵</label><input class="acct-input" type="email"></div>
              <div class="acct-field full"><label>備註</label><textarea class="acct-textarea" placeholder="可補充與審批人核對所需資料"></textarea></div>
            </div>
          </section>
          <section class="work-card">
            <div class="work-card-title">所需文件</div>
            <div class="binding-doc-grid">
              <div v-for="doc in bindingDocs" :key="doc.title" class="binding-doc-card">
                <div class="binding-doc-title">{{ doc.title }}</div>
                <div class="binding-doc-desc">{{ doc.desc }}</div>
                <input class="acct-file" type="file">
              </div>
            </div>
            <div class="work-bind-actions">
              <button type="button" class="work-action">提交審批</button>
            </div>
          </section>
          <section class="work-card">
            <div class="work-card-title">申請狀態</div>
            <div class="binding-status-list">
              <div v-for="item in bindingStatusList" :key="item.title" class="binding-status-item">
                <div>
                  <div class="binding-status-title">{{ item.title }}</div>
                  <div class="binding-status-meta">{{ item.meta }}</div>
                </div>
                <span class="work-chip" :class="item.chipType">{{ item.chip }}</span>
              </div>
            </div>
          </section>
        </div>

        <!-- 2.4 AJO 錢包 -->
        <div v-show="activePanel === 'profile-wallet'" class="work-panel on">
          <section class="work-hero">
            <div>
              <div class="work-kicker">Wallet</div>
              <h2 class="work-title">AJO 錢包</h2>
              <p class="work-desc">查看 AJO Point 累計、使用與最近流水。</p>
            </div>
          </section>
          <section class="work-grid">
            <div class="work-card">
              <div class="work-stat">{{ walletStats.earned }}</div>
              <div class="work-stat-label">累計獲得 AJO Point</div>
            </div>
            <div class="work-card">
              <div class="work-stat">{{ walletStats.used }}</div>
              <div class="work-stat-label">累計使用 AJO Point</div>
            </div>
            <div class="work-card">
              <div class="work-card-title">充值積分</div>
              <div class="work-card-sub">充值接口暫未開放，現階段可透過觀看廣告或營運發放取得積分。</div>
            </div>
          </section>
          <section class="work-card">
            <div class="work-card-title">最近流水</div>
            <table class="work-table">
              <thead><tr><th>類型</th><th>來源</th><th>積分</th><th>日期</th></tr></thead>
              <tbody>
                <tr v-for="(item, idx) in walletTransactions" :key="idx">
                  <td>{{ item.type }}</td>
                  <td>{{ item.source }}</td>
                  <td>{{ item.points }}</td>
                  <td>{{ item.date }}</td>
                </tr>
              </tbody>
            </table>
          </section>
        </div>

        <!-- 2.5 聊天 -->
        <div v-show="activePanel === 'profile-chat'" class="work-panel on">
          <section class="work-hero">
            <div>
              <div class="work-kicker">Chat</div>
              <h2 class="work-title">聊天</h2>
              <p class="work-desc">查看與通知、買家、租客或管理員的會話。</p>
            </div>
          </section>
          <section class="profile-chat-shell">
            <aside class="profile-chat-list" aria-label="會話列表">
              <div class="profile-chat-list-head">
                <div class="work-card-title">會話</div>
                <span class="profile-chat-count">{{ chatThreads.length }}</span>
              </div>
              <button
                v-for="thread in chatThreads"
                :key="thread.id"
                type="button"
                class="profile-chat-thread"
                :class="{ active: thread.active }"
              >
                <span class="profile-chat-avatar">{{ thread.avatar }}</span>
                <span class="profile-chat-thread-main">
                  <span class="profile-chat-thread-top">
                    <span class="profile-chat-thread-name">{{ thread.name }}</span>
                    <span class="profile-chat-time">{{ thread.time }}</span>
                  </span>
                  <span class="profile-chat-preview">{{ thread.preview }}</span>
                  <span v-if="thread.unread" class="profile-chat-unread">未讀</span>
                </span>
              </button>
            </aside>
            <section class="profile-chat-window" aria-label="聊天窗口">
              <header class="profile-chat-header">
                <div>
                  <div class="profile-chat-name">租務查詢</div>
                  <div class="profile-chat-meta">關聯樓盤：佐敦 高臨 代理盤</div>
                </div>
                <span class="work-chip brand">待回覆</span>
              </header>
              <div class="profile-chat-messages">
                <template v-for="(group, idx) in chatMessages" :key="idx">
                  <div class="profile-chat-day">{{ group.day }}</div>
                  <div
                    v-for="(row, ridx) in group.rows"
                    :key="ridx"
                    class="profile-chat-row"
                    :class="{ sent: row.sent }"
                  >
                    <div class="profile-chat-bubble">
                      {{ row.text }}
                      <span class="profile-chat-bubble-time">{{ row.time }}</span>
                    </div>
                  </div>
                </template>
              </div>
              <div class="profile-chat-compose">
                <input class="profile-chat-input" placeholder="輸入訊息">
                <button type="button" class="profile-chat-send">送出</button>
              </div>
            </section>
          </section>
        </div>

        <!-- 2.6 我的樓盤 -->
        <div v-show="activePanel === 'profile-properties'" class="work-panel on">
          <section class="work-hero">
            <div>
              <div class="work-kicker">Properties</div>
              <h2 class="work-title">我的樓盤</h2>
              <p class="work-desc">管理已發布、草稿、上架中和已下架樓盤。</p>
            </div>
            <button type="button" class="work-action">新增發布</button>
          </section>
          <section class="work-card">
            <div class="work-toolbar">
              <input class="work-search" placeholder="搜尋標題、屋苑、地點">
              <button type="button" class="work-mini-btn primary">搜尋</button>
            </div>
            <table class="work-table">
              <thead><tr><th>樓盤</th><th>地區</th><th>價格</th><th>狀態</th><th>操作</th></tr></thead>
              <tbody>
                <tr v-for="(item, idx) in propertyListings" :key="idx">
                  <td>{{ item.title }}</td>
                  <td>{{ item.area }}</td>
                  <td>{{ item.price }}</td>
                  <td><span class="work-chip" :class="item.chipType">{{ item.status }}</span></td>
                  <td>
                    <div class="work-table-actions">
                      <button type="button" class="work-mini-btn">編輯</button>
                      <button type="button" class="work-mini-btn">公開頁</button>
                      <button v-if="item.chipType === 'good'" type="button" class="work-mini-btn">下架</button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </section>
        </div>

        <!-- 2.7 我的住宅 -->
        <div v-show="activePanel === 'profile-homes'" class="work-panel on">
          <section class="work-hero">
            <div>
              <div class="work-kicker">Residential</div>
              <h2 class="work-title">我的住宅</h2>
              <p class="work-desc">查看已綁定住宅與服務式住宅申請。</p>
            </div>
          </section>
          <section class="work-card">
            <div class="work-card-title">住宅列表</div>
            <table class="work-table">
              <thead><tr><th>住宅</th><th>身份</th><th>狀態</th><th>更新時間</th></tr></thead>
              <tbody>
                <tr v-for="(item, idx) in homeListings" :key="idx">
                  <td>{{ item.name }}</td>
                  <td>{{ item.identity }}</td>
                  <td><span class="work-chip" :class="item.chipType">{{ item.status }}</span></td>
                  <td>{{ item.updatedAt }}</td>
                </tr>
              </tbody>
            </table>
          </section>
        </div>

        <!-- 2.8 我的家具 -->
        <div v-show="activePanel === 'profile-furniture'" class="work-panel on">
          <section class="work-hero">
            <div>
              <div class="work-kicker">Furniture</div>
              <h2 class="work-title">我的家具</h2>
              <p class="work-desc">管理二手家私發布與交易狀態。</p>
            </div>
            <button type="button" class="work-action">新增家私</button>
          </section>
          <section class="work-card">
            <div class="work-card-title">家具列表</div>
            <table class="work-table">
              <thead><tr><th>商品</th><th>分類</th><th>價格</th><th>狀態</th><th>操作</th></tr></thead>
              <tbody>
                <tr v-for="(item, idx) in furnitureListings" :key="idx">
                  <td>{{ item.name }}</td>
                  <td>{{ item.category }}</td>
                  <td>{{ item.price }}</td>
                  <td><span class="work-chip" :class="item.chipType">{{ item.status }}</span></td>
                  <td><button type="button" class="work-mini-btn">編輯</button></td>
                </tr>
              </tbody>
            </table>
          </section>
        </div>

        <!-- 2.9 我的收藏 -->
        <div v-show="activePanel === 'profile-saved'" class="work-panel on">
          <section class="work-card">
            <div class="work-card-title">收藏列表</div>
            <table class="work-table">
              <thead><tr><th>項目</th><th>分類</th><th>價格</th><th>收藏時間</th></tr></thead>
              <tbody>
                <tr v-for="(item, idx) in savedItems" :key="idx">
                  <td>{{ item.name }}</td>
                  <td>{{ item.category }}</td>
                  <td>{{ item.price }}</td>
                  <td>{{ item.savedAt }}</td>
                </tr>
              </tbody>
            </table>
          </section>
        </div>

        <!-- 2.10 交易訂單 -->
        <div v-show="activePanel === 'profile-orders'" class="work-panel on">
          <section class="work-hero">
            <div>
              <div class="work-kicker">Orders</div>
              <h2 class="work-title">交易訂單</h2>
              <p class="work-desc">查看近期交易與訂單狀態。</p>
            </div>
          </section>
          <section class="work-card">
            <div class="work-card-title">訂單列表</div>
            <table class="work-table">
              <thead><tr><th>訂單</th><th>類型</th><th>金額</th><th>狀態</th></tr></thead>
              <tbody>
                <tr v-for="(item, idx) in orderListings" :key="idx">
                  <td>{{ item.id }}</td>
                  <td>{{ item.type }}</td>
                  <td>{{ item.amount }}</td>
                  <td><span class="work-chip" :class="item.chipType">{{ item.status }}</span></td>
                </tr>
              </tbody>
            </table>
          </section>
        </div>
      </div>
    </div>
  </main>
</template>

<style scoped>
/* 1. 頁面容器 */
.account-my-page {
  width: 100%;
  min-height: calc(100vh - var(--nav-h, 52px));
  background: rgb(var(--color-surface-2));
  color: rgb(var(--color-text));
}

/* 2. 雙欄佈局 */
.work-shell {
  display: grid;
  grid-template-columns: 240px minmax(0, 1fr);
  gap: 16px;
  max-width: 1180px;
  margin: 0 auto;
  padding: 12px 24px 16px;
}

/* 3. 左側導覽 */
.work-sidebar {
  position: sticky;
  top: 72px;
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

.work-nav-item.on,
.work-nav-item:focus-visible {
  background: transparent;
  color: rgb(var(--color-primary));
  font-weight: 700;
  outline: none;
}

.work-nav-item.on:hover,
.work-nav-item:focus-visible:hover {
  background: rgb(var(--color-primary-soft));
}

.work-nav-item.on::after,
.work-nav-item:focus-visible::after {
  transform: scaleX(1);
}

/* 4. 右側主內容 */
.work-main {
  display: grid;
  gap: 12px;
  min-width: 0;
  align-content: start;
}

.work-panel {
  display: grid;
  gap: 14px;
  align-content: start;
}

/* 5. 區段標題 */
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
  padding: 0 0 12px;
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

/* 6. 通用按鈕 */
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

.work-compact-action {
  min-height: 38px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 700;
  padding: 9px 14px;
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

/* 7. 卡片 */
.work-card {
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 16px;
}

.work-card--spaced {
  margin-top: 14px;
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

/* 8. 帳號頂欄 */
.work-account-topbar {
  display: flex;
  width: 100%;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  flex-wrap: wrap;
}

.work-account-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
}

.work-account-user {
  display: flex;
  align-items: center;
  gap: 10px;
}

.work-account-avatar {
  display: flex;
  width: 42px;
  height: 42px;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: rgb(var(--color-primary-soft));
  border: 1px solid #eadbd4;
  color: rgb(var(--color-primary));
  font-size: 14px;
  font-weight: 800;
}

.work-account-name {
  color: rgb(var(--color-text));
  font-size: 14px;
  font-weight: 800;
  line-height: 1.2;
}

.work-account-sub {
  margin-top: 3px;
  color: rgb(var(--color-ink-3));
  font-size: 12px;
  font-weight: 700;
}

.work-account-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.work-account-btn {
  display: inline-flex;
  min-height: 34px;
  align-items: center;
  justify-content: center;
  border: 1px solid rgb(var(--color-border));
  border-radius: 6px;
  background: #fff;
  color: rgb(var(--color-text));
  cursor: pointer;
  font-family: inherit;
  font-size: 12px;
  font-weight: 700;
  padding: 0 12px;
  line-height: 1;
}

.work-account-btn.outline {
  border-color: #f6c7b3;
  box-shadow: inset 0 0 0 1px #f6c7b3;
}

.work-account-btn.primary {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: #fff;
}

/* 9. 帳號資料網格 */
.work-account-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  overflow: hidden;
  margin-top: 12px;
}

.work-account-card {
  background: rgb(var(--color-surface));
  overflow: hidden;
}

.work-account-card:first-child {
  border-right: 1px solid rgb(var(--color-border));
}

.work-account-card .work-card-title {
  margin: 0;
  padding: 14px 16px 12px;
  color: rgb(var(--color-primary));
  font-size: 14px;
  font-weight: 800;
}

.work-account-field {
  display: grid;
  grid-template-columns: 150px minmax(0, 1fr);
  gap: 12px;
  border-top: 1px solid rgb(var(--color-surface-3));
  padding: 10px 16px;
}

.work-account-field:first-child {
  border-top: 0;
}

.work-account-field span {
  color: rgb(var(--color-ink-3));
  font-size: 12px;
  font-weight: 700;
  line-height: 1.5;
}

.work-account-field strong {
  color: rgb(var(--color-text));
  font-size: 13px;
  font-weight: 700;
  line-height: 1.5;
  word-break: break-word;
}

.work-role-strip {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  padding: 10px 16px;
  border-top: 1px solid rgb(var(--color-surface-3));
}

.work-role-badge {
  display: inline-flex;
  align-items: center;
  border-radius: 4px;
  background: rgb(var(--color-primary));
  color: #fff;
  font-size: 12px;
  font-weight: 800;
  padding: 7px 10px;
}

.work-role-tag {
  display: inline-flex;
  align-items: center;
  border: 1px solid rgb(var(--color-border));
  border-radius: 4px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-ink-2));
  font-size: 12px;
  font-weight: 700;
  padding: 7px 10px;
}

/* 10. 綁定單位 */
.work-bind-wrap {
  display: grid;
  gap: 12px;
}

.work-bind-current {
  color: rgb(var(--color-text));
  font-size: 14px;
  font-weight: 800;
}

.work-bind-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.work-bind-field {
  display: grid;
  gap: 8px;
}

.work-bind-label {
  color: rgb(var(--color-ink-3));
  font-size: 12px;
  font-weight: 700;
}

.work-bind-select {
  width: 100%;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: #fff;
  color: rgb(var(--color-text));
  font-family: inherit;
  font-size: 14px;
  font-weight: 700;
  padding: 14px 16px;
}

.work-bind-actions {
  display: flex;
  justify-content: flex-end;
}

/* 11. iSmart 卡片 */
.work-ismart-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  margin-top: 12px;
}

.work-ismart-card {
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 14px;
}

.work-ismart-card .work-card-sub {
  font-size: 13px;
  font-weight: 800;
  color: rgb(var(--color-text));
  margin-bottom: 6px;
}

.work-ismart-card .work-row {
  padding: 10px 0;
}

.work-ismart-card .work-row > div {
  display: grid;
  grid-template-columns: 150px minmax(0, 1fr);
  gap: 12px;
  width: 100%;
}

.work-ismart-card .work-row strong {
  font-size: 12px;
  font-weight: 700;
  color: rgb(var(--color-ink-3));
  line-height: 1.5;
}

.work-ismart-card .work-row span {
  margin-top: 0;
  font-size: 13px;
  font-weight: 700;
  color: rgb(var(--color-text));
  line-height: 1.5;
  word-break: break-word;
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

/* 12. 統計 */
.work-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
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

/* 13. 狀態標籤 */
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

/* 14. 表格 */
.work-table-wrap {
  overflow: auto;
}

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

.work-table-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.work-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.work-search {
  flex: 1;
  border: 1px solid rgb(var(--color-border));
  border-radius: 6px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
  font-family: inherit;
  font-size: 13px;
  padding: 10px 12px;
}

/* 15. 授權副戶 */
.work-subaccount-group {
  display: grid;
  gap: 14px;
}

.work-subaccount-card {
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 14px;
}

.work-subaccount-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.work-subaccount-name {
  color: rgb(var(--color-text));
  font-size: 18px;
  font-weight: 700;
}

.work-subaccount-empty {
  color: rgb(var(--color-ink-3));
  font-size: 12px;
  font-weight: 600;
  text-align: center;
  padding: 18px 12px;
}

/* 16. 提示設定 */
.work-setting-box {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 14px;
  margin-top: 12px;
}

.work-setting-check {
  display: flex;
  align-items: center;
  gap: 10px;
  color: rgb(var(--color-text));
  font-size: 14px;
  font-weight: 700;
}

.work-setting-check input {
  width: 16px;
  height: 16px;
  accent-color: rgb(var(--color-primary));
}

/* 17. 物業綁定 */
.binding-flow {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 10px;
}

.binding-step {
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 14px;
}

.binding-step-num {
  display: flex;
  width: 26px;
  height: 26px;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: rgb(var(--color-primary));
  color: #fff;
  font-size: 12px;
  font-weight: 800;
  margin-bottom: 10px;
}

.binding-step-title {
  color: rgb(var(--color-text));
  font-size: 13px;
  font-weight: 800;
  margin-bottom: 5px;
}

.binding-step-desc {
  color: rgb(var(--color-ink-3));
  font-size: 12px;
  line-height: 1.6;
}

.binding-platform-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.binding-platform-card {
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: #fff;
  color: rgb(var(--color-text));
  cursor: pointer;
  font-family: inherit;
  text-align: left;
  padding: 16px;
  transition: border-color 0.15s, background 0.15s, box-shadow 0.15s;
}

.binding-platform-card:hover {
  border-color: rgb(var(--color-brand-mid));
  box-shadow: var(--shadow-soft);
}

.binding-platform-card.on {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary-soft));
}

.binding-platform-card strong {
  display: block;
  color: rgb(var(--color-text));
  font-size: 14px;
  font-weight: 800;
  margin-bottom: 6px;
}

.binding-platform-card span {
  display: block;
  color: rgb(var(--color-ink-3));
  font-size: 12px;
  line-height: 1.6;
}

.binding-review-box {
  display: grid;
  gap: 6px;
  margin-top: 12px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface-2));
  padding: 12px 14px;
}

.binding-review-box strong {
  color: rgb(var(--color-text));
  font-size: 12px;
  font-weight: 800;
}

.binding-review-box span {
  color: rgb(var(--color-ink-3));
  font-size: 12px;
  line-height: 1.6;
}

.binding-platform-panel {
  display: block;
}

.binding-subtitle {
  color: rgb(var(--color-text));
  font-size: 12px;
  font-weight: 800;
  margin: 0 0 10px;
}

.binding-subtitle:not(:first-child) {
  margin-top: 16px;
}

.binding-common-grid {
  margin-top: 14px;
}

.binding-doc-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.binding-doc-card {
  border: 1px dashed rgb(var(--color-border-2));
  border-radius: 8px;
  background: rgb(var(--color-surface-2));
  padding: 14px;
}

.binding-doc-title {
  color: rgb(var(--color-text));
  font-size: 13px;
  font-weight: 800;
  margin-bottom: 5px;
}

.binding-doc-desc {
  color: rgb(var(--color-ink-3));
  font-size: 12px;
  line-height: 1.6;
  margin-bottom: 10px;
}

.binding-status-list {
  display: grid;
  gap: 10px;
}

.binding-status-item {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 12px;
  align-items: center;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: #fff;
  padding: 12px 14px;
}

.binding-status-title {
  color: rgb(var(--color-text));
  font-size: 13px;
  font-weight: 800;
}

.binding-status-meta {
  color: rgb(var(--color-ink-3));
  font-size: 12px;
  margin-top: 4px;
}

/* 18. 表單 */
.acct-form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.acct-field {
  display: grid;
  gap: 7px;
}

.acct-field.full {
  grid-column: 1 / -1;
}

.acct-field label {
  color: rgb(var(--color-text));
  font-size: 12px;
  font-weight: 800;
}

.acct-input,
.acct-select,
.acct-textarea {
  width: 100%;
  border: 1px solid rgb(var(--color-border));
  border-radius: 6px;
  background: #fff;
  color: rgb(var(--color-text));
  font-family: inherit;
  font-size: 13px;
  font-weight: 600;
  padding: 10px 12px;
}

.acct-textarea {
  min-height: 86px;
  resize: vertical;
  line-height: 1.6;
}

.acct-file {
  flex: 1;
  min-width: 260px;
  border: 1px dashed rgb(var(--color-border-2));
  border-radius: 6px;
  background: rgb(var(--color-surface-2));
  padding: 12px;
  color: rgb(var(--color-ink-3));
  font-size: 12px;
  font-weight: 700;
}

/* 19. 聊天 */
.profile-chat-shell {
  display: grid;
  grid-template-columns: 282px minmax(0, 1fr);
  min-height: 510px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  overflow: hidden;
}

.profile-chat-list {
  border-right: 1px solid rgb(var(--color-border));
  background: rgb(var(--color-surface));
}

.profile-chat-list-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  border-bottom: 1px solid rgb(var(--color-border));
  padding: 14px 16px;
}

.profile-chat-list-head .work-card-title {
  margin-bottom: 0;
}

.profile-chat-count {
  border-radius: 999px;
  background: rgb(var(--color-primary-soft));
  color: rgb(var(--color-primary));
  font-size: 11px;
  font-weight: 800;
  padding: 4px 8px;
}

.profile-chat-thread {
  display: grid;
  width: 100%;
  grid-template-columns: 38px minmax(0, 1fr);
  gap: 10px;
  border: 0;
  border-bottom: 1px solid rgb(var(--color-surface-3));
  background: transparent;
  color: rgb(var(--color-text));
  cursor: pointer;
  font-family: inherit;
  padding: 13px 14px;
  text-align: left;
}

.profile-chat-thread:hover,
.profile-chat-thread.active {
  background: rgb(var(--color-primary-soft));
}

.profile-chat-avatar {
  display: flex;
  width: 38px;
  height: 38px;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: rgb(var(--color-surface-2));
  color: rgb(var(--color-ink-2));
  font-size: 13px;
  font-weight: 800;
}

.profile-chat-thread.active .profile-chat-avatar {
  background: rgb(var(--color-primary));
  color: #fff;
}

.profile-chat-thread-main {
  min-width: 0;
}

.profile-chat-thread-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  margin-bottom: 4px;
}

.profile-chat-thread-name {
  min-width: 0;
  overflow: hidden;
  color: rgb(var(--color-text));
  font-size: 13px;
  font-weight: 800;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.profile-chat-time {
  color: rgb(var(--color-ink-4));
  font-size: 10px;
  font-weight: 700;
  white-space: nowrap;
}

.profile-chat-preview {
  overflow: hidden;
  color: rgb(var(--color-ink-3));
  font-size: 12px;
  line-height: 1.45;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.profile-chat-unread {
  display: inline-flex;
  margin-top: 7px;
  border-radius: 999px;
  background: rgb(var(--color-primary));
  color: #fff;
  font-size: 10px;
  font-weight: 800;
  padding: 3px 7px;
}

.profile-chat-window {
  display: flex;
  min-width: 0;
  flex-direction: column;
  background: rgb(var(--color-surface));
}

.profile-chat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  border-bottom: 1px solid rgb(var(--color-border));
  padding: 14px 18px;
}

.profile-chat-name {
  color: rgb(var(--color-text));
  font-size: 15px;
  font-weight: 800;
}

.profile-chat-meta {
  margin-top: 4px;
  color: rgb(var(--color-ink-3));
  font-size: 12px;
}

.profile-chat-messages {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 10px;
  overflow: auto;
  background: rgb(var(--color-surface-2));
  padding: 18px;
}

.profile-chat-day {
  align-self: center;
  border-radius: 999px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-ink-3));
  font-size: 10px;
  font-weight: 700;
  padding: 4px 9px;
}

.profile-chat-row {
  display: flex;
}

.profile-chat-row.sent {
  justify-content: flex-end;
}

.profile-chat-bubble {
  max-width: min(72%, 520px);
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-ink-2));
  font-size: 13px;
  line-height: 1.6;
  padding: 10px 12px;
}

.profile-chat-row.sent .profile-chat-bubble {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: #fff;
}

.profile-chat-bubble-time {
  display: block;
  margin-top: 5px;
  color: rgb(var(--color-ink-4));
  font-size: 10px;
  font-weight: 700;
}

.profile-chat-row.sent .profile-chat-bubble-time {
  color: rgba(255, 255, 255, 0.72);
}

.profile-chat-compose {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 10px;
  border-top: 1px solid rgb(var(--color-border));
  padding: 12px;
  background: rgb(var(--color-surface));
}

.profile-chat-input {
  width: 100%;
  border: 1px solid rgb(var(--color-border));
  border-radius: 6px;
  background: #fff;
  color: rgb(var(--color-text));
  font-family: inherit;
  font-size: 13px;
  padding: 10px 12px;
}

.profile-chat-input:focus {
  border-color: rgb(var(--color-primary));
  box-shadow: 0 0 0 3px rgba(240, 90, 0, 0.1);
  outline: none;
}

.profile-chat-send {
  border: 0;
  border-radius: 6px;
  background: rgb(var(--color-primary));
  color: #fff;
  cursor: pointer;
  font-family: inherit;
  font-size: 13px;
  font-weight: 800;
  padding: 0 16px;
}

/* 20. 響應式 */
@media (max-width: 900px) {
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

  .work-grid,
  .work-ismart-grid,
  .work-account-grid,
  .work-bind-grid,
  .binding-flow,
  .binding-platform-grid,
  .binding-doc-grid {
    grid-template-columns: 1fr;
  }

  .work-account-card:first-child {
    border-right: 0;
    border-bottom: 1px solid rgb(var(--color-border));
  }

  .profile-chat-shell {
    grid-template-columns: 1fr;
    min-height: 0;
  }

  .profile-chat-list {
    border-right: 0;
    border-bottom: 1px solid rgb(var(--color-border));
  }

  .profile-chat-window {
    min-height: 440px;
  }

  .acct-form-grid {
    grid-template-columns: 1fr;
  }
}
</style>
