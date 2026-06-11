<!--
 * 首頁 scroll grid 元件。
 * 1. 以傾斜 mosaic grid 呈現目前聚焦中的首頁模組。
 * 2. 隨首頁 active section 切換卡片內容、色調與位移。
-->
<script setup lang="ts">
import { computed } from 'vue';

import type { HomeGridTile, HomeModuleCode, HomeModuleDefinition } from '@/pages/home/home';

interface HomeScrollGridProps {
  modules: HomeModuleDefinition[];
  activeCode: HomeModuleCode;
}

const props = defineProps<HomeScrollGridProps>();

const tileMotionMap: Record<HomeModuleCode, Array<{ x: number; y: number; rotate: number; scale: number }>> = {
  secondhand: [
    { x: -10, y: -14, rotate: -4, scale: 1.02 },
    { x: 12, y: -8, rotate: 2, scale: 1 },
    { x: -18, y: 8, rotate: -2, scale: 1.04 },
    { x: 8, y: 14, rotate: 4, scale: 0.99 },
    { x: 18, y: -10, rotate: 3, scale: 1.01 },
    { x: -12, y: 10, rotate: -3, scale: 1.01 },
    { x: 10, y: 6, rotate: 1, scale: 0.98 },
    { x: -4, y: -6, rotate: -1, scale: 1 },
  ],
  property_sale: [
    { x: -4, y: -8, rotate: -2, scale: 1 },
    { x: 16, y: -14, rotate: 2, scale: 1.02 },
    { x: -10, y: 14, rotate: -4, scale: 1.03 },
    { x: 6, y: 16, rotate: 3, scale: 1 },
    { x: 18, y: -2, rotate: 1, scale: 0.98 },
    { x: -18, y: 6, rotate: -2, scale: 1.01 },
    { x: 12, y: 10, rotate: 2, scale: 1 },
    { x: -8, y: -10, rotate: -1, scale: 0.99 },
  ],
  serviced_apartment: [
    { x: -8, y: -12, rotate: -3, scale: 1.01 },
    { x: 10, y: -6, rotate: 3, scale: 1 },
    { x: -16, y: 12, rotate: -2, scale: 1.04 },
    { x: 12, y: 18, rotate: 4, scale: 0.99 },
    { x: 20, y: -8, rotate: 2, scale: 1.02 },
    { x: -10, y: 12, rotate: -2, scale: 1 },
    { x: 6, y: 8, rotate: 1, scale: 0.99 },
    { x: -2, y: -8, rotate: -1, scale: 1 },
  ],
};

const activeModule = computed(
  () => props.modules.find((item) => item.code === props.activeCode) ?? props.modules[0],
);

const sizeClassMap: Record<HomeGridTile['size'], string> = {
  hero: 'scroll-grid-tile--hero',
  wide: 'scroll-grid-tile--wide',
  tall: 'scroll-grid-tile--tall',
  square: 'scroll-grid-tile--square',
};

// 1. resolveTileClassName 組合單一 grid 區塊樣式
const resolveTileClassName = (tile: HomeGridTile) => [
  'scroll-grid-tile',
  sizeClassMap[tile.size],
];

// 2. resolveTileStyle 輸出單一卡片動畫位移
const resolveTileStyle = (index: number) => {
  const motion = tileMotionMap[activeModule.value.code][index] ?? {
    x: 0,
    y: 0,
    rotate: 0,
    scale: 1,
  };

  return {
    '--tile-shift-x': `${motion.x}px`,
    '--tile-shift-y': `${motion.y}px`,
    '--tile-rotate': `${motion.rotate}deg`,
    '--tile-scale': String(motion.scale),
    '--tile-delay': `${index * 42}ms`,
  };
};
</script>

<template>
  <section
    class="scroll-grid-shell"
    :data-tone="activeModule.tone"
  >
    <div class="scroll-grid-stage">
      <div class="scroll-grid-stage__aurora" />
      <div class="scroll-grid-stage__shadow" />

      <div class="scroll-grid-plane">
        <article
          v-for="(tile, index) in activeModule.gridTiles"
          :key="`${activeModule.code}-${tile.id}`"
          :class="resolveTileClassName(tile)"
          :data-tone="tile.tone"
          :style="resolveTileStyle(index)"
        >
          <p class="scroll-grid-tile__eyebrow">
            {{ tile.eyebrow }}
          </p>
          <div class="scroll-grid-tile__body">
            <p class="scroll-grid-tile__title">
              {{ tile.title }}
            </p>
            <p class="scroll-grid-tile__value">
              {{ tile.value }}
            </p>
          </div>
        </article>
      </div>
    </div>

    <div class="panel-surface scroll-grid-caption p-5 sm:p-6">
      <div class="flex items-start justify-between gap-4">
        <div>
          <p class="text-kicker">
            {{ activeModule.gridEyebrow }}
          </p>
          <h3 class="scroll-grid-caption__title mt-3">
            {{ activeModule.displayTitle }}
          </h3>
        </div>

        <span
          class="scroll-grid-caption__status"
          :class="activeModule.isLive ? 'scroll-grid-caption__status--live' : 'scroll-grid-caption__status--planned'"
        >
          {{ activeModule.availabilityLabel }}
        </span>
      </div>

      <p class="mt-4 text-sm leading-7 text-text-muted">
        {{ activeModule.gridDescription }}
      </p>

      <div class="mt-5 flex flex-wrap gap-2">
        <span
          v-for="chip in activeModule.chips"
          :key="chip"
          class="scroll-grid-caption__chip"
        >
          {{ chip }}
        </span>
      </div>
    </div>
  </section>
</template>

<style scoped>
.scroll-grid-shell {
  display: grid;
  gap: 1.5rem;
}

.scroll-grid-stage {
  position: relative;
  overflow: hidden;
  min-height: 33rem;
  padding: 1.75rem;
  border: 1px solid rgb(var(--color-border) / 0.72);
  border-radius: 3px;
  background: rgb(var(--color-surface));
  box-shadow: 0 10px 24px rgb(70 53 37 / 0.1);
}

.scroll-grid-stage__aurora,
.scroll-grid-stage__shadow {
  position: absolute;
  inset: auto;
  pointer-events: none;
}

.scroll-grid-stage__aurora {
  display: none;
  top: -5rem;
  right: -3rem;
  width: 16rem;
  height: 16rem;
  border-radius: 999px;
  filter: blur(18px);
  opacity: 0.5;
}

.scroll-grid-stage__shadow {
  display: none;
  right: 1.25rem;
  bottom: 1rem;
  left: 1.25rem;
  height: 5rem;
  border-radius: 999px;
  background: transparent;
  transform: translateY(1.5rem) scaleX(0.88);
}

.scroll-grid-shell[data-tone='copper'] .scroll-grid-stage__aurora {
  background: rgb(var(--color-primary-soft));
}

.scroll-grid-shell[data-tone='slate'] .scroll-grid-stage__aurora {
  background: rgb(var(--color-surface-muted));
}

.scroll-grid-shell[data-tone='sage'] .scroll-grid-stage__aurora {
  background: rgb(var(--color-surface-raised));
}

.scroll-grid-plane {
  position: relative;
  z-index: 1;
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  grid-auto-rows: 4.85rem;
  gap: 0.95rem;
  transform: perspective(1400px) rotateX(57deg) rotateZ(-24deg) translateY(1rem);
  transform-style: preserve-3d;
}

.scroll-grid-tile {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 0.85rem;
  min-height: 0;
  padding: 1rem;
  border: 1px solid rgb(var(--color-border) / 0.68);
  border-radius: 3px;
  color: rgb(var(--color-text));
  box-shadow: 0 10px 24px rgb(70 53 37 / 0.1);
  transform:
    translate3d(var(--tile-shift-x), var(--tile-shift-y), 0)
    rotate(var(--tile-rotate))
    scale(var(--tile-scale));
  transition:
    transform 0.58s cubic-bezier(0.22, 1, 0.36, 1),
    background 0.42s ease,
    border-color 0.42s ease,
    color 0.42s ease;
  transition-delay: var(--tile-delay);
}

.scroll-grid-tile--hero {
  grid-column: span 2;
  grid-row: span 2;
}

.scroll-grid-tile--wide {
  grid-column: span 2;
}

.scroll-grid-tile--tall {
  grid-row: span 2;
}

.scroll-grid-tile--square {
  grid-column: span 1;
}

.scroll-grid-tile[data-tone='accent'] {
  background: rgb(var(--color-primary-soft));
}

.scroll-grid-tile[data-tone='soft'] {
  background: rgb(var(--color-surface-raised));
}

.scroll-grid-tile[data-tone='contrast'] {
  border-color: rgb(var(--color-text) / 0.18);
  background:
    linear-gradient(180deg, rgb(78 64 52 / 0.96), rgb(49 39 31 / 0.94));
  color: rgb(255 248 242);
}

.scroll-grid-tile[data-tone='neutral'] {
  background: rgb(var(--color-surface));
}

.scroll-grid-tile__eyebrow {
  margin: 0;
  font-size: 0.7rem;
  font-weight: 700;
  line-height: 1.3;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  opacity: 0.72;
}

.scroll-grid-tile__body {
  display: grid;
  gap: 0.35rem;
}

.scroll-grid-tile__title {
  margin: 0;
  font-family: var(--font-display);
  font-size: 1.35rem;
  line-height: 1.05;
}

.scroll-grid-tile__value {
  margin: 0;
  font-size: 0.82rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  opacity: 0.8;
}

.scroll-grid-caption {
  position: relative;
  z-index: 1;
}

.scroll-grid-caption__title {
  margin: 0;
  font-family: var(--font-display);
  font-size: 1.7rem;
  line-height: 1.14;
  color: rgb(var(--color-text));
}

.scroll-grid-caption__status {
  display: inline-flex;
  align-items: center;
  padding: 0.55rem 0.8rem;
  border-radius: 999px;
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  white-space: nowrap;
}

.scroll-grid-caption__status--live {
  background: rgb(var(--color-success) / 0.12);
  color: rgb(var(--color-success));
}

.scroll-grid-caption__status--planned {
  background: rgb(var(--color-text) / 0.08);
  color: rgb(var(--color-text-muted));
}

.scroll-grid-caption__chip {
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

@media (max-width: 1023px) {
  .scroll-grid-stage {
    min-height: 27rem;
    padding: 1.15rem;
  }

  .scroll-grid-plane {
    grid-auto-rows: 4.25rem;
    gap: 0.75rem;
    transform: none;
  }

  .scroll-grid-tile {
    transition-duration: 0.36s;
  }
}

@media (max-width: 639px) {
  .scroll-grid-plane {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    grid-auto-rows: 4.5rem;
  }

  .scroll-grid-tile--hero,
  .scroll-grid-tile--wide,
  .scroll-grid-tile--tall {
    grid-column: span 2;
    grid-row: span 1;
  }

  .scroll-grid-stage {
    min-height: auto;
  }
}
</style>
