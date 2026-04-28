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
  <div class="relative hidden h-full overflow-hidden lg:block">
    <img
      :src="props.hero.src"
      :alt="t('auth.hero.imageAlt')"
      class="h-full w-full object-cover object-center"
    >
    <div class="absolute inset-0 bg-gradient-to-br from-black/55 via-black/20 to-transparent" />
    <div class="absolute inset-0 flex flex-col justify-between p-12 pt-32 xl:p-16 xl:pt-36">
      <div>
        <p class="text-xs font-semibold uppercase tracking-[0.32em] text-white/70">
          AJO Living
        </p>
        <h1 class="mt-5 max-w-xl font-display text-5xl leading-tight text-white drop-shadow xl:text-6xl">
          {{ t('auth.hero.title') }}
        </h1>
        <p class="mt-5 max-w-md text-lg leading-8 text-white/82 drop-shadow">
          {{ t('auth.hero.description') }}
        </p>
      </div>

      <div class="flex flex-wrap items-end justify-between gap-6 text-white/80">
        <div class="grid grid-cols-[auto_1px_auto] items-center gap-6">
          <div>
            <p class="text-xs font-bold uppercase tracking-[0.24em] text-white/60">
              {{ t('auth.hero.locationLabel') }}
            </p>
            <p class="mt-2 text-sm font-medium">
              {{ props.hero.location }}
            </p>
          </div>
          <div class="h-10 w-px bg-white/25" />
          <div>
            <p class="text-xs font-bold uppercase tracking-[0.24em] text-white/60">
              {{ t('auth.hero.photoLabel') }}
            </p>
            <p class="mt-2 text-sm font-medium">
              {{ props.hero.author }}
            </p>
          </div>
        </div>
        <p class="rounded-full border border-white/20 bg-white/10 px-4 py-2 text-xs font-semibold uppercase tracking-[0.22em] backdrop-blur">
          {{ t('auth.hero.established') }}
        </p>
      </div>
    </div>
  </div>
</template>
