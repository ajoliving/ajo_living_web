<!--
 * 登入頁主入口。
 * 1. 高保真還原 HTML 設計稿雙欄版面：左側品牌視覺區，右側登入表單。
 * 2. 使用靜態 mock 資料，移除所有 API 呼叫。
 * 3. 支援登入/註冊模式切換、Tab 切換、記住我、路由跳轉。
-->
<script setup lang="ts">
import { computed, ref } from 'vue';
import { useRouter } from 'vue-router';

const router = useRouter();

// 1. 模式狀態：login / register
const authMode = ref<'login' | 'register'>('login');

// 2. 登入 Tab 狀態：phone / username / email
const loginTab = ref<'phone' | 'username' | 'email'>('username');

// 3. 表單欄位 mock 資料
const loginIdentifier = ref('');
const loginPassword = ref('');
const rememberMe = ref(true);

const registerUsername = ref('');
const registerPhone = ref('');
const registerPassword = ref('');
const registerBuilding = ref('');
const registerFloor = ref('');
const registerUnit = ref('');
const registerEmail = ref('');

// 4. 大廈、樓層、單位 mock 選項
const buildingOptions = ['康睦庭園第二座', '海景花園'];
const floorOptions = ['02', '08'];
const unitOptions = ['D', 'A'];

// 5. 登入 Tab 對應的標籤與佔位符
const loginTabMeta = computed(() => {
  if (loginTab.value === 'phone') {
    return { label: '用戶名 / 手機 / 電郵', placeholder: '例如 +852 6123 4567' };
  }
  if (loginTab.value === 'email') {
    return { label: '用戶名 / 手機 / 電郵', placeholder: 'name@example.com' };
  }
  return { label: '用戶名 / 手機 / 電郵', placeholder: 'patrick' };
});

// 6. 切換登入/註冊模式
const setAuthMode = (mode: 'login' | 'register') => {
  authMode.value = mode;
};

// 7. 切換登入 Tab
const setLoginTab = (tab: 'phone' | 'username' | 'email') => {
  loginTab.value = tab;
};

// 8. 登入提交：mock 行為，跳轉首頁
const handleLoginSubmit = () => {
  router.push({ name: 'Home' });
};

// 9. 註冊提交：mock 行為，跳回首頁
const handleRegisterSubmit = () => {
  router.push({ name: 'Home' });
};
</script>

<template>
  <div class="login-page">
    <div class="auth-shell">
      <!-- 1. 左側品牌視覺區 -->
      <section class="auth-visual">
        <div class="auth-visual-main">
          <div class="auth-kicker">AJO LIVING</div>
          <div class="auth-title">歡迎回家</div>
          <div class="auth-desc">開啟您的理想社區生活，管理發佈、聯絡與鄰里交易。</div>
        </div>
        <div class="auth-visual-foot">
          <div>地點<br>Hong Kong</div>
          <div>圖片<br>AJO Living</div>
          <div>創立於 2024</div>
        </div>
      </section>

      <!-- 2. 右側登入表單區 -->
      <section class="auth-panel">
        <div class="auth-card">
          <!-- 2.1 登入表單 -->
          <template v-if="authMode === 'login'">
            <div class="auth-heading">住戶登入</div>
            <div class="auth-copy">可使用用戶名、手機號碼或電郵配合密碼登入。</div>

            <div class="auth-form on">
              <div class="ltabs">
                <button
                  type="button"
                  class="ltab"
                  :class="{ on: loginTab === 'phone' }"
                  @click="setLoginTab('phone')"
                >
                  手機登入
                </button>
                <button
                  type="button"
                  class="ltab"
                  :class="{ on: loginTab === 'username' }"
                  @click="setLoginTab('username')"
                >
                  用戶名
                </button>
                <button
                  type="button"
                  class="ltab"
                  :class="{ on: loginTab === 'email' }"
                  @click="setLoginTab('email')"
                >
                  郵箱登入
                </button>
              </div>

              <div class="auth-field">
                <label class="auth-label">{{ loginTabMeta.label }}</label>
                <div class="auth-input-wrap">
                  <input
                    v-model="loginIdentifier"
                    class="linput"
                    :placeholder="loginTabMeta.placeholder"
                  />
                </div>
              </div>

              <div class="auth-field">
                <label class="auth-label">密碼</label>
                <div class="auth-input-wrap">
                  <input
                    v-model="loginPassword"
                    type="password"
                    class="linput"
                    placeholder="請輸入密碼"
                  />
                </div>
              </div>

              <label class="auth-remember">
                <input v-model="rememberMe" type="checkbox" />
                記住我
              </label>

              <button class="auth-submit" type="button" @click="handleLoginSubmit">
                登入
              </button>

              <div class="auth-foot">
                還沒有帳戶？
                <button class="auth-link" type="button" @click="setAuthMode('register')">
                  註冊
                </button>
              </div>
            </div>
          </template>

          <!-- 2.2 註冊表單 -->
          <template v-else>
            <div class="auth-heading">建立帳戶</div>
            <div class="auth-copy">填寫基本資料完成註冊，加入 AJO Living 社區。</div>

            <div class="auth-form on">
              <div class="ltabs">
                <button type="button" class="ltab on">新用戶</button>
                <button type="button" class="ltab">已有 ismart 帳戶</button>
              </div>

              <div class="auth-field">
                <label class="auth-label">用戶名</label>
                <div class="auth-input-wrap">
                  <input v-model="registerUsername" class="linput" placeholder="請輸入用戶名" />
                </div>
              </div>

              <div class="auth-field">
                <label class="auth-label">手機號碼</label>
                <div class="auth-input-wrap">
                  <input v-model="registerPhone" class="linput" placeholder="例如 +852 6123 4567" />
                </div>
              </div>

              <div class="auth-field">
                <label class="auth-label">密碼</label>
                <div class="auth-input-wrap">
                  <input
                    v-model="registerPassword"
                    type="password"
                    class="linput"
                    placeholder="請設定登入密碼"
                  />
                </div>
              </div>

              <div class="auth-field">
                <label class="auth-label">大廈</label>
                <div class="auth-input-wrap">
                  <select v-model="registerBuilding" class="linput">
                    <option value="">請選擇大廈</option>
                    <option v-for="b in buildingOptions" :key="b" :value="b">{{ b }}</option>
                  </select>
                </div>
              </div>

              <div class="auth-grid-2">
                <div class="auth-field">
                  <label class="auth-label">樓層</label>
                  <div class="auth-input-wrap">
                    <select v-model="registerFloor" class="linput">
                      <option value="">請選擇樓層</option>
                      <option v-for="f in floorOptions" :key="f" :value="f">{{ f }}</option>
                    </select>
                  </div>
                </div>
                <div class="auth-field">
                  <label class="auth-label">單位</label>
                  <div class="auth-input-wrap">
                    <select v-model="registerUnit" class="linput">
                      <option value="">請選擇單位</option>
                      <option v-for="u in unitOptions" :key="u" :value="u">{{ u }}</option>
                    </select>
                  </div>
                </div>
              </div>

              <div class="auth-field">
                <label class="auth-label">郵箱</label>
                <div class="auth-input-wrap">
                  <input v-model="registerEmail" class="linput" placeholder="name@example.com" />
                </div>
              </div>

              <button class="auth-submit" type="button" @click="handleRegisterSubmit">
                建立帳戶
              </button>

              <div class="auth-foot">
                已經有帳戶？
                <button class="auth-link" type="button" @click="setAuthMode('login')">
                  登入
                </button>
              </div>
            </div>
          </template>
        </div>
      </section>
    </div>
  </div>
</template>

<style scoped>
/*
 * 登入頁樣式。
 * 1. auth-shell 雙欄版面：左側品牌視覺，右側表單。
 * 2. auth-visual 暗色漸層背景，含 kicker / title / desc / foot。
 * 3. auth-panel 置中表單卡片，含 tabs / field / input / submit / foot。
 * 4. 響應式：768px 以下改為單欄，480px 以下縮減 padding。
 */

.login-page {
  width: 100%;
  min-height: calc(100svh - 48px);
  background: rgb(var(--color-surface));
}

/* 1. 雙欄版面 */
.auth-shell {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 520px;
  min-height: calc(100svh - 48px);
  background: rgb(var(--color-surface));
}

/* 2. 左側品牌視覺區 */
.auth-visual {
  position: relative;
  display: flex;
  min-height: 640px;
  flex-direction: column;
  justify-content: space-between;
  overflow: hidden;
  background: linear-gradient(160deg, #26313a 0%, #536570 52%, #1f2529 100%);
  padding: 58px;
  color: #fff;
}

.auth-visual::before {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(90deg, rgba(0, 0, 0, 0.62), rgba(0, 0, 0, 0.18));
  z-index: 1;
}

.auth-visual-main,
.auth-visual-foot {
  position: relative;
  z-index: 2;
}

.auth-kicker {
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 4px;
  color: rgba(255, 255, 255, 0.72);
}

.auth-title {
  margin-top: 26px;
  font-family: var(--font-display);
  font-size: 50px;
  font-weight: 400;
  line-height: 1.1;
  color: #fff;
}

.auth-desc {
  max-width: 520px;
  margin-top: 18px;
  color: rgba(255, 255, 255, 0.75);
  font-size: 16px;
  font-weight: 600;
  line-height: 1.8;
}

.auth-visual-foot {
  display: flex;
  justify-content: space-between;
  gap: 18px;
  color: rgba(255, 255, 255, 0.8);
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.6px;
  line-height: 1.6;
}

/* 3. 右側表單區 */
.auth-panel {
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgb(var(--color-surface));
  padding: 24px 32px;
}

.auth-card {
  width: 100%;
  max-width: 430px;
}

.auth-heading {
  font-family: var(--font-display);
  font-size: 30px;
  font-weight: 400;
  color: rgb(var(--color-text));
  line-height: 1.2;
}

.auth-copy {
  margin-top: 8px;
  color: rgb(var(--color-ink-3));
  font-size: 13px;
  font-weight: 600;
  line-height: 1.6;
}

.auth-form {
  margin-top: 14px;
}

/* 3.1 Tab 列 */
.ltabs {
  display: flex;
  border-bottom: 1px solid rgb(var(--color-border));
  margin-bottom: 18px;
}

.ltab {
  font-size: 14px;
  font-weight: 800;
  padding: 8px 0;
  margin-right: 18px;
  cursor: pointer;
  color: rgb(var(--color-ink-3));
  border: none;
  border-bottom: 2px solid transparent;
  background: none;
  font-family: inherit;
  transition: color 0.15s ease, border-color 0.15s ease;
}

.ltab:hover {
  color: rgb(var(--color-ink-2));
}

.ltab.on {
  color: rgb(var(--color-primary));
  border-bottom: 2px solid rgb(var(--color-primary));
}

/* 3.2 欄位 */
.auth-field {
  margin-bottom: 12px;
}

.auth-label {
  display: block;
  margin-bottom: 7px;
  color: rgb(var(--color-ink-3));
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0.8px;
}

.auth-input-wrap {
  position: relative;
}

.linput {
  width: 100%;
  height: 46px;
  border: 1px solid rgb(var(--color-border));
  padding: 10px 12px;
  font-size: 13px;
  font-family: inherit;
  outline: none;
  border-radius: 2px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
  transition: border-color 0.15s ease;
}

.linput:focus {
  border-color: rgb(var(--color-primary));
}

.linput::placeholder {
  color: rgb(var(--color-ink-4));
}

select.linput {
  appearance: none;
  cursor: pointer;
}

.auth-grid-2 {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

/* 3.3 記住我 */
.auth-remember {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 2px 0 10px;
  color: rgb(var(--color-ink-2));
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
}

.auth-remember input {
  cursor: pointer;
  accent-color: rgb(var(--color-primary));
}

/* 3.4 提交按鈕 */
.auth-submit {
  width: 100%;
  margin-top: 6px;
  background: rgb(var(--color-primary));
  color: #fff;
  border: 0;
  border-radius: 6px;
  cursor: pointer;
  font-family: inherit;
  font-size: 14px;
  font-weight: 800;
  letter-spacing: 2px;
  padding: 13px;
  transition: background 0.15s ease;
}

.auth-submit:hover {
  background: rgb(var(--color-brand-dark));
}

/* 3.5 底部連結 */
.auth-foot {
  margin-top: 16px;
  border-top: 1px solid rgb(var(--color-border));
  padding-top: 14px;
  text-align: center;
  color: rgb(var(--color-ink-3));
  font-size: 13px;
  font-weight: 700;
}

.auth-link {
  border: 0;
  background: transparent;
  color: rgb(var(--color-primary));
  cursor: pointer;
  font-family: inherit;
  font-size: 13px;
  font-weight: 800;
  padding: 0;
}

.auth-link:hover {
  color: rgb(var(--color-brand-dark));
}

/* 4. 響應式 */
@media (max-width: 1024px) {
  .auth-shell {
    grid-template-columns: minmax(0, 1fr) 440px;
  }
}

@media (max-width: 768px) {
  .auth-shell {
    grid-template-columns: 1fr;
  }

  .auth-visual {
    min-height: 320px;
    padding: 34px 26px;
  }

  .auth-title {
    font-size: 38px;
  }

  .auth-desc {
    font-size: 14px;
  }

  .auth-panel {
    padding: 20px 20px;
  }
}

@media (max-width: 480px) {
  .auth-panel {
    padding: 16px 14px;
  }

  .auth-card {
    max-width: none;
  }

  .auth-grid-2 {
    grid-template-columns: 1fr;
  }
}
</style>
