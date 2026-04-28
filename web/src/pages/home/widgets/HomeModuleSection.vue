<!--
 * 首頁模組說明區塊。
 * 1. 呈現單一首頁模組的定位、指標、亮點與操作入口。
 * 2. 配合 scroll active 狀態輸出強弱視覺層級。
-->
<script setup lang="ts">
import { RouterLink } from 'vue-router';

import type { HomeModuleDefinition } from '@/pages/home/home';
import AppIcon from '@/shared/components/base/AppIcon.vue';
import BaseButton from '@/shared/components/base/BaseButton.vue';

interface HomeModuleSectionProps {
  module: HomeModuleDefinition;
  active: boolean;
}

const props = defineProps<HomeModuleSectionProps>();
</script>

<template>
  <article
    :id="`home-module-${props.module.code}`"
    class="home-module panel-surface p-6 sm:p-8"
    :class="props.active ? 'home-module--active' : 'home-module--idle'"
    :data-home-module="props.module.code"
    :data-tone="props.module.tone"
  >
    <div class="home-module__orb" />

    <div class="relative space-y-8">
      <div class="flex flex-col gap-6 xl:flex-row xl:items-start xl:justify-between">
        <div class="max-w-3xl">
          <div class="flex flex-wrap items-center gap-3">
            <p class="text-kicker">
              {{ props.module.kicker }}
            </p>
            <span
              class="home-module__status"
              :class="props.module.isLive ? 'home-module__status--live' : 'home-module__status--planned'"
            >
              {{ props.module.availabilityLabel }}
            </span>
          </div>

          <h3 class="home-module__title mt-4">
            {{ props.module.title }}
          </h3>

          <p class="mt-4 max-w-3xl text-base leading-8 text-text-muted">
            {{ props.module.description }}
          </p>
        </div>

        <div class="home-module__icon-shell">
          <span class="home-module__index">{{ props.module.index }}</span>
          <div class="home-module__icon">
            <AppIcon
              :name="props.module.iconName"
              :size="24"
            />
          </div>
        </div>
      </div>

      <div class="grid gap-5 lg:grid-cols-[1fr_auto] lg:items-end">
        <div class="grid gap-3 sm:grid-cols-3">
          <div
            v-for="metric in props.module.metrics"
            :key="metric.label"
            class="home-module__metric"
          >
            <p class="home-module__metric-value">
              {{ metric.value }}
            </p>
            <p class="home-module__metric-label">
              {{ metric.label }}
            </p>
          </div>
        </div>

        <div class="flex flex-col gap-3 sm:flex-row sm:items-center lg:justify-end">
          <RouterLink
            v-if="props.module.primaryAction"
            :to="props.module.primaryAction.to"
          >
            <BaseButton
              size="lg"
              :variant="props.module.primaryAction.variant"
            >
              {{ props.module.primaryAction.label }}
            </BaseButton>
          </RouterLink>

          <RouterLink
            v-if="props.module.secondaryAction"
            :to="props.module.secondaryAction.to"
          >
            <BaseButton
              size="lg"
              :variant="props.module.secondaryAction.variant"
            >
              {{ props.module.secondaryAction.label }}
            </BaseButton>
          </RouterLink>

          <div
            v-if="!props.module.primaryAction"
            class="home-module__coming-soon"
          >
            <span class="home-module__coming-soon-dot" />
            <span>{{ props.module.availabilityLabel }}</span>
          </div>
        </div>
      </div>
    </div>
  </article>
</template>

<style scoped>
.home-module {
  position: relative;
  overflow: hidden;
  transition:
    transform 0.36s ease,
    box-shadow 0.36s ease,
    border-color 0.36s ease;
}

.home-module--active {
  transform: translateY(-4px);
  box-shadow: 0 30px 70px rgb(80 61 43 / 0.16);
}

.home-module--idle {
  opacity: 0.96;
}

.home-module__orb {
  position: absolute;
  top: -6rem;
  right: -5rem;
  width: 15rem;
  height: 15rem;
  border-radius: 999px;
  filter: blur(18px);
  opacity: 0.52;
}

.home-module[data-tone='copper'] .home-module__orb {
  background: radial-gradient(circle, rgb(193 146 104 / 0.3), transparent 70%);
}

.home-module[data-tone='slate'] .home-module__orb {
  background: radial-gradient(circle, rgb(120 136 152 / 0.28), transparent 70%);
}

.home-module[data-tone='sage'] .home-module__orb {
  background: radial-gradient(circle, rgb(119 153 134 / 0.24), transparent 70%);
}

.home-module__title {
  margin: 0;
  font-family: var(--font-display);
  font-size: clamp(2rem, 4vw, 2.9rem);
  line-height: 1.05;
  color: rgb(var(--color-text));
}

.home-module__icon-shell {
  display: flex;
  align-items: center;
  gap: 1rem;
  align-self: flex-start;
}

.home-module__index {
  font-family: var(--font-display);
  font-size: 4rem;
  line-height: 1;
  color: rgb(var(--color-text) / 0.12);
}

.home-module__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 3.5rem;
  height: 3.5rem;
  border: 1px solid rgb(var(--color-border) / 0.72);
  border-radius: 1.35rem;
  background: rgb(var(--color-surface) / 0.88);
  color: rgb(var(--color-text));
}

.home-module__status {
  display: inline-flex;
  align-items: center;
  padding: 0.5rem 0.8rem;
  border-radius: 999px;
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
}

.home-module__status--live {
  background: rgb(var(--color-success) / 0.12);
  color: rgb(var(--color-success));
}

.home-module__status--planned {
  background: rgb(var(--color-text) / 0.08);
  color: rgb(var(--color-text-muted));
}

.home-module__metric {
  padding: 1rem 1rem 1.1rem;
  border: 1px solid rgb(var(--color-border) / 0.68);
  border-radius: 1.4rem;
  background:
    linear-gradient(180deg, rgb(var(--color-surface) / 0.94), rgb(var(--color-surface-raised) / 0.88));
}

.home-module__metric-value {
  margin: 0;
  font-family: var(--font-display);
  font-size: 1.7rem;
  line-height: 1;
  color: rgb(var(--color-text));
}

.home-module__metric-label {
  margin: 0.75rem 0 0;
  font-size: 0.82rem;
  font-weight: 600;
  line-height: 1.5;
  color: rgb(var(--color-text-muted));
}

.home-module__chip {
  display: inline-flex;
  align-items: center;
  padding: 0.55rem 0.85rem;
  border: 1px solid rgb(var(--color-border) / 0.7);
  border-radius: 999px;
  background: rgb(var(--color-surface) / 0.88);
  font-size: 0.78rem;
  font-weight: 600;
  color: rgb(var(--color-text-muted));
}

.home-module__highlight {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 0.85rem;
  align-items: start;
  padding: 1rem 1rem 1.05rem;
  border: 1px solid rgb(var(--color-border) / 0.62);
  border-radius: 1.35rem;
  background: rgb(var(--color-surface) / 0.86);
}

.home-module__highlight-marker {
  width: 0.65rem;
  height: 0.65rem;
  margin-top: 0.55rem;
  border-radius: 999px;
  background: rgb(var(--color-primary));
  box-shadow: 0 0 0 5px rgb(var(--color-primary) / 0.12);
}

.home-module__coming-soon {
  display: inline-flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.95rem 1rem;
  border-radius: 999px;
  background: rgb(var(--color-text) / 0.07);
  font-size: 0.85rem;
  font-weight: 700;
  color: rgb(var(--color-text-muted));
}

.home-module__coming-soon-dot {
  width: 0.55rem;
  height: 0.55rem;
  border-radius: 999px;
  background: rgb(var(--color-warning));
}

@media (max-width: 767px) {
  .home-module__index {
    font-size: 3.1rem;
  }

  .home-module__title {
    font-size: 2rem;
  }
}
</style>
