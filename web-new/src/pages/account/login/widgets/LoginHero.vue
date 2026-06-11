<!--
 * 登入頁主視覺。
 * 1. 展示登入頁背景圖片、標題與說明文字。
 * 2. 呈現圖片來源與地點資訊。
-->
<script setup lang="ts">
import { useI18n } from 'vue-i18n';

import type { LoginHeroImage } from '../login';

interface LoginHeroProps {
  hero: LoginHeroImage;
}

const props = defineProps<LoginHeroProps>();
const { t } = useI18n();
</script>

<template>
  <div class="login-hero hidden h-full overflow-hidden lg:block">
    <img
      :src="props.hero.src"
      :alt="t('auth.hero.imageAlt')"
      class="login-hero__image"
    >
    <div class="login-hero__shade" />
    <div class="login-hero__content">
      <div>
        <p class="login-hero__eyebrow">
          AJO Living
        </p>
        <h1>
          {{ t('auth.hero.title') }}
        </h1>
        <p class="login-hero__description">
          {{ t('auth.hero.description') }}
        </p>
      </div>

      <div class="login-hero__meta">
        <div class="login-hero__meta-grid">
          <div>
            <p>
              {{ t('auth.hero.locationLabel') }}
            </p>
            <span>
              {{ props.hero.location }}
            </span>
          </div>
          <div class="login-hero__divider" />
          <div>
            <p>
              {{ t('auth.hero.photoLabel') }}
            </p>
            <span>
              {{ props.hero.author }}
            </span>
          </div>
        </div>
        <p class="login-hero__badge">
          {{ t('auth.hero.established') }}
        </p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-hero {
  position: relative;
}

.login-hero__image,
.login-hero__shade,
.login-hero__content {
  position: absolute;
  inset: 0;
}

.login-hero__image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: center;
}

.login-hero__shade {
  background: linear-gradient(90deg, rgb(0 0 0 / 0.72), rgb(0 0 0 / 0.18));
}

.login-hero__content {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 72px 48px 40px;
}

.login-hero__eyebrow,
.login-hero__meta p,
.login-hero__badge {
  color: rgb(255 255 255 / 0.68);
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.24em;
  text-transform: uppercase;
}

.login-hero h1 {
  max-width: 560px;
  margin-top: 18px;
  color: #ffffff;
  font-family: var(--font-display);
  font-size: 46px;
  font-weight: 400;
  line-height: 1.14;
}

.login-hero__description {
  max-width: 420px;
  margin-top: 14px;
  color: rgb(255 255 255 / 0.78);
  font-size: 14px;
  line-height: 1.8;
}

.login-hero__meta {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: 16px;
  color: rgb(255 255 255 / 0.82);
}

.login-hero__meta-grid {
  display: grid;
  grid-template-columns: auto 1px auto;
  align-items: center;
  gap: 18px;
}

.login-hero__meta span {
  display: block;
  margin-top: 6px;
  font-size: 12px;
  font-weight: 500;
}

.login-hero__divider {
  width: 1px;
  height: 32px;
  background: rgb(255 255 255 / 0.24);
}

.login-hero__badge {
  border: 1px solid rgb(255 255 255 / 0.28);
  border-radius: 2px;
  padding: 7px 10px;
}
</style>
