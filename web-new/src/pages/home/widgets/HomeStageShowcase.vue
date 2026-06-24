<!--
 * 首頁首屏與舞台區。
 * 1. 保留原有首頁舞台動畫。
 * 2. 在首屏下方加入圖片輪播圖。
 * 3. 移除首頁主標上方的額外標籤與 01 02 03 導覽列。
-->
<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { gsap } from 'gsap';
import { ScrollTrigger } from 'gsap/ScrollTrigger';

import type {
  HomeLandingContent,
  HomeModuleCode,
  HomeModuleDefinition,
} from '@/pages/home/home';

gsap.registerPlugin(ScrollTrigger);

/*
 * 1. Props 定義
 */
interface HomeStageShowcaseProps {
  content: HomeLandingContent;
}

const props = defineProps<HomeStageShowcaseProps>();

const emit = defineEmits<{
  (e: 'active-change', code: HomeModuleCode): void;
}>();

/*
 * 2. 首屏輪播狀態
 */
const modules = computed(() => props.content.modules);
const carouselRoot = ref<HTMLElement | null>(null);

interface HomeCarouselSlide {
  key: string;
  module: HomeModuleDefinition;
  imageUrl: string;
  imageAlt: string;
}

const CAROUSEL_GAP = 28;
const CAROUSEL_MAX_ROTATION = 28;
const CAROUSEL_MAX_DEPTH = 140;
const CAROUSEL_MIN_SCALE = 0.92;
const CAROUSEL_SCALE_RANGE = 0.1;
const CAROUSEL_FRICTION = 0.9;
const CAROUSEL_WHEEL_SENS = 0.6;
const CAROUSEL_DRAG_SENS = 1;

let carouselCards: HTMLElement[] = [];
let carouselFrameId: number | null = null;
let carouselLastFrame = 0;
let carouselScrollX = 0;
let carouselVelocityX = 0;
let carouselStep = 0;
let carouselTrack = 0;
let carouselViewportHalf = 0;
let isCarouselDragging = false;
let dragLastX = 0;
let dragLastTime = 0;
let dragLastDelta = 0;

const modulo = (value: number, base: number) => ((value % base) + base) % base;

// 2.1 建立重複輪播資料，避免只有三張卡片時軌道過短
const carouselSlides = computed<HomeCarouselSlide[]>(() =>
  Array.from({ length: 4 }, (_, repeatIndex) =>
    carouselBaseSlides.value.map((slide, slideIndex) => ({
      ...slide,
      key: `${slide.key}-${repeatIndex}-${slideIndex}`,
    })),
  ).flat(),
);

// 2.2 建立首頁輪播基礎資料
const carouselBaseSlides = computed<HomeCarouselSlide[]>(() => {
  if (props.content.carouselImages.length > 0) {
    return props.content.carouselImages.map((image, index) => {
      const module = modules.value[index % Math.max(modules.value.length, 1)] ?? modules.value[0];

      return {
        key: image.id,
        module,
        imageUrl: image.url,
        imageAlt: image.alt,
      };
    });
  }

  return modules.value.map((module) => ({
    key: module.code,
    module,
    imageUrl: module.carouselImagePath,
    imageAlt: module.displayTitle,
  }));
});

const bindCarouselCards = async () => {
  await nextTick();
  carouselCards = Array.from(
    carouselRoot.value?.querySelectorAll<HTMLElement>('.home-carousel__slide') ?? [],
  );
};

const measureCarousel = () => {
  if (!carouselCards.length) {
    return;
  }

  const sample = carouselCards[0];
  const cardRect = sample.getBoundingClientRect();
  carouselStep = cardRect.width + CAROUSEL_GAP;
  carouselTrack = carouselStep * carouselCards.length;
  carouselViewportHalf = window.innerWidth * 0.5;
};

const updateCarouselTransforms = () => {
  if (!carouselCards.length || !carouselTrack) {
    return;
  }

  const halfTrack = carouselTrack / 2;
  const wrappedPositions = new Array<number>(carouselCards.length);
  let closestIndex = 0;
  let closestDistance = Number.POSITIVE_INFINITY;

  carouselCards.forEach((_, index) => {
    let position = index * carouselStep - carouselScrollX;

    if (position < -halfTrack) {
      position += carouselTrack;
    }
    if (position > halfTrack) {
      position -= carouselTrack;
    }

    wrappedPositions[index] = position;
    const distance = Math.abs(position);
    if (distance < closestDistance) {
      closestDistance = distance;
      closestIndex = index;
    }
  });

  const previousIndex = (closestIndex - 1 + carouselCards.length) % carouselCards.length;
  const nextIndex = (closestIndex + 1) % carouselCards.length;

  carouselCards.forEach((card, index) => {
    const position = wrappedPositions[index] ?? 0;
    const norm = Math.max(-1, Math.min(1, position / carouselViewportHalf));
    const inverseNorm = 1 - Math.abs(norm);
    const rotationY = -norm * CAROUSEL_MAX_ROTATION;
    const depthZ = inverseNorm * CAROUSEL_MAX_DEPTH;
    const scale = CAROUSEL_MIN_SCALE + inverseNorm * CAROUSEL_SCALE_RANGE;
    const isCore =
      index === closestIndex || index === previousIndex || index === nextIndex;
    const blur = isCore ? 0 : 2 * Math.pow(Math.abs(norm), 1.1);

    card.style.transform = `translate3d(${position}px,-50%,${depthZ}px) rotateY(${rotationY}deg) scale(${scale})`;
    card.style.filter = `blur(${blur.toFixed(2)}px)`;
    card.style.opacity = '1';
    card.style.zIndex = String(1000 + Math.round(depthZ));
  });

  if (carouselSlides.value.length > 0) {
    const nextModule = carouselSlides.value[closestIndex]?.module;
    if (nextModule) {
      emit('active-change', nextModule.code);
    }
  }
};

const animateCarousel = (timestamp: number) => {
  const deltaSeconds = carouselLastFrame
    ? (timestamp - carouselLastFrame) / 1000
    : 0;
  carouselLastFrame = timestamp;

  if (carouselTrack > 0) {
    carouselScrollX = modulo(
      carouselScrollX + carouselVelocityX * deltaSeconds,
      carouselTrack,
    );

    const decay = Math.pow(CAROUSEL_FRICTION, deltaSeconds * 60);
    carouselVelocityX *= decay;
    if (Math.abs(carouselVelocityX) < 0.02) {
      carouselVelocityX = 0;
    }
  }

  updateCarouselTransforms();
  carouselFrameId = window.requestAnimationFrame(animateCarousel);
};

const startCarousel = () => {
  if (carouselFrameId) {
    window.cancelAnimationFrame(carouselFrameId);
  }
  carouselLastFrame = 0;
  carouselFrameId = window.requestAnimationFrame(animateCarousel);
};

const stopCarousel = () => {
  if (carouselFrameId) {
    window.cancelAnimationFrame(carouselFrameId);
    carouselFrameId = null;
  }
};

/*
 * 2.1 以滑鼠滾輪注入慣性速度
 */
const handleCarouselWheel = (event: WheelEvent) => {
  if (!carouselTrack) {
    return;
  }

  event.preventDefault();
  const dominantDelta = Math.abs(event.deltaX) > Math.abs(event.deltaY)
    ? event.deltaX
    : event.deltaY;
  carouselVelocityX += dominantDelta * CAROUSEL_WHEEL_SENS * 20;
};

/*
 * 2.2 啟動拖拽操作
 */
const handleCarouselPointerDown = (event: PointerEvent) => {
  if (!carouselTrack || !carouselRoot.value) {
    return;
  }

  isCarouselDragging = true;
  dragLastX = event.clientX;
  dragLastTime = performance.now();
  dragLastDelta = 0;
  try {
    carouselRoot.value.setPointerCapture(event.pointerId);
  } catch {
    // ignore pointer capture failure
  }
  carouselRoot.value.classList.add('dragging');
};

/*
 * 2.3 拖拽時更新輪播位置與速度
 */
const handleCarouselPointerMove = (event: PointerEvent) => {
  if (!isCarouselDragging || !carouselTrack) {
    return;
  }

  const now = performance.now();
  const deltaX = event.clientX - dragLastX;
  const deltaTime = Math.max(1, now - dragLastTime) / 1000;

  carouselScrollX = modulo(carouselScrollX - deltaX * CAROUSEL_DRAG_SENS, carouselTrack);
  dragLastDelta = deltaX / deltaTime;
  dragLastX = event.clientX;
  dragLastTime = now;
};

/*
 * 2.4 結束拖拽並保留慣性
 */
const releaseCarouselPointer = (event: PointerEvent) => {
  if (!isCarouselDragging || !carouselRoot.value) {
    return;
  }

  isCarouselDragging = false;
  try {
    carouselRoot.value.releasePointerCapture(event.pointerId);
  } catch {
    // ignore pointer release failure
  }
  carouselRoot.value.classList.remove('dragging');
  carouselVelocityX = -dragLastDelta * CAROUSEL_DRAG_SENS;
};

/*
 * 3. 舞台動畫狀態
 */
let master: gsap.core.Timeline | null = null;
let resizeTimer = 0;
let activeStageCode: HomeModuleCode | null = null;

const svgNS = 'http://www.w3.org/2000/svg';
const GRID_ROWS = 0;
const STAGE_TRIGGER_START = 'top 74%';
const STAGE_TRIGGER_END = 'bottom bottom';
let blindsSets: SVGRectElement[][] = [];

/*
 * 4. 舞台網格欄數
 */
function getGridCols() {
  if (window.innerWidth <= 599) return 6;
  if (window.innerWidth <= 1024) return 10;
  return 14;
}

/*
 * 5. 建立遮罩方格
 */
function createBlinds(groupId: string) {
  const group = document.getElementById(groupId);
  if (!group) return null;

  group.innerHTML = '';

  const width = window.innerWidth;
  const height = window.innerHeight;
  const viewBoxWidth = 100;
  const viewBoxHeight = (height / width) * 100;
  const cols = getGridCols();
  const rows = GRID_ROWS || Math.round(cols * (viewBoxHeight / viewBoxWidth));
  const cellWidth = viewBoxWidth / cols;
  const cellHeight = viewBoxHeight / rows;

  const cells: SVGRectElement[] = [];

  for (let row = 0; row < rows; row += 1) {
    for (let col = 0; col < cols; col += 1) {
      const rect = document.createElementNS(svgNS, 'rect');
      rect.setAttribute('x', String(col * cellWidth));
      rect.setAttribute('y', String(row * cellHeight));
      rect.setAttribute('width', String(cellWidth));
      rect.setAttribute('height', String(cellHeight));
      rect.setAttribute('fill', 'white');
      rect.setAttribute('shape-rendering', 'crispEdges');
      rect.setAttribute('opacity', '0');
      group.appendChild(rect);
      cells.push(rect);
    }
  }

  return cells;
}

/*
 * 6. 網格開啟動畫
 */
function openBlinds(cells: SVGRectElement[]) {
  const shuffled = gsap.utils.shuffle([...cells]);

  return gsap.timeline().to(shuffled, {
    opacity: 1,
    duration: 1,
    ease: 'power3.out',
    stagger: { each: 0.02 },
  });
}

/*
 * 7. 文字進出動畫
 */
function textIn(element: Element) {
  return gsap.to(element, {
    clipPath: 'inset(0% 0% 0% 0%)',
    y: 0,
    duration: 2.6,
    ease: 'expo.out',
  });
}

function textOut(element: Element) {
  return gsap.to(element, {
    clipPath: 'inset(0% 0% 100% 0%)',
    y: 0,
    duration: 2,
    ease: 'power2.inOut',
  });
}

/*
 * 8. 重新計算舞台版面
 */
function updateLayout() {
  const width = window.innerWidth;
  const height = window.innerHeight;
  const viewBoxWidth = 100;
  const viewBoxHeight = (height / width) * 100;

  const layers = document.querySelectorAll<SVGSVGElement>('.layer');
  blindsSets = [];

  layers.forEach((svg) => {
    svg.setAttribute('viewBox', `0 0 ${viewBoxWidth} ${viewBoxHeight}`);

    const maskRect = svg.querySelector('mask rect');
    if (maskRect) {
      maskRect.setAttribute('width', String(viewBoxWidth));
      maskRect.setAttribute('height', String(viewBoxHeight));
    }

    const image = svg.querySelector('image');
    if (image) {
      image.setAttribute('width', String(viewBoxWidth));
      image.setAttribute('height', String(viewBoxHeight));
    }

    const blindId = (svg.querySelector('g[id^="blinds"]') as SVGGElement | null)?.id;
    if (blindId) {
      const blinds = createBlinds(blindId);
      if (blinds) {
        blindsSets.push(blinds);
      }
    }
  });

  buildMasterTimeline();
}

/*
 * 9. 建立舞台主時間軸
 */
function buildMasterTimeline() {
  if (master) {
    master.kill();
  }

  const texts = gsap.utils.toArray<Element>('.txt');

  master = gsap.timeline({
    scrollTrigger: {
      trigger: '.stage',
      start: STAGE_TRIGGER_START,
      end: STAGE_TRIGGER_END,
      scrub: 2.5,
      anticipatePin: 1,
      invalidateOnRefresh: true,
    },
  });

  blindsSets.forEach((cells, index) => {
    master?.add(openBlinds(cells));
    if (texts[index]) {
      master?.add(textIn(texts[index]), '-=0.3');
      master?.add(textOut(texts[index]), '+=0.8');
    }
  });
}

/*
 * 10. 初始化進度條
 */
function initProgressBar() {
  const fills = gsap.utils.toArray<HTMLElement>('.progress-bar .fill');

  ScrollTrigger.create({
    trigger: '.stage',
    start: STAGE_TRIGGER_START,
    end: STAGE_TRIGGER_END,
    scrub: 0.3,
    onUpdate: (self) => {
      const totalSteps = fills.length;

      fills.forEach((fill, index) => {
        let progress = (self.progress - index / totalSteps) * totalSteps;
        progress = Math.max(0, Math.min(1, progress));
        fill.style.width = `${progress * 100}%`;
      });
    },
  });
}

/*
 * 11. 同步舞台 active module
 */
function initActiveModuleTracker() {
  ScrollTrigger.create({
    trigger: '.stage',
    start: STAGE_TRIGGER_START,
    end: STAGE_TRIGGER_END,
    onUpdate: (self) => {
      const progress = self.progress;
      const index = Math.min(Math.floor(progress * modules.value.length), modules.value.length - 1);
      const code = modules.value[index]?.code;

      if (code && code !== activeStageCode) {
        activeStageCode = code;
        emit('active-change', code);
      }
    },
  });
}

/*
 * 12. resize 防抖
 */
function handleResize() {
  clearTimeout(resizeTimer);
  resizeTimer = window.setTimeout(() => {
    measureCarousel();
    updateCarouselTransforms();
    ScrollTrigger.refresh();
    updateLayout();
  }, 250);
}

/*
 * 13. 初始化與清理
 */
onMounted(async () => {
  await bindCarouselCards();
  measureCarousel();
  updateCarouselTransforms();
  startCarousel();
  carouselRoot.value?.classList.add('carousel-mode');
  carouselRoot.value?.addEventListener('wheel', handleCarouselWheel, { passive: false });
  carouselRoot.value?.addEventListener('pointerdown', handleCarouselPointerDown);
  carouselRoot.value?.addEventListener('pointermove', handleCarouselPointerMove);
  carouselRoot.value?.addEventListener('pointerup', releaseCarouselPointer);
  carouselRoot.value?.addEventListener('pointercancel', releaseCarouselPointer);
  updateLayout();
  initProgressBar();
  initActiveModuleTracker();
  window.addEventListener('resize', handleResize);
});

watch(
  () => carouselSlides.value.map((slide) => slide.key).join('|'),
  async () => {
    await bindCarouselCards();
    measureCarousel();
    updateCarouselTransforms();
  },
);

onBeforeUnmount(() => {
  stopCarousel();
  carouselRoot.value?.removeEventListener('wheel', handleCarouselWheel);
  carouselRoot.value?.removeEventListener('pointerdown', handleCarouselPointerDown);
  carouselRoot.value?.removeEventListener('pointermove', handleCarouselPointerMove);
  carouselRoot.value?.removeEventListener('pointerup', releaseCarouselPointer);
  carouselRoot.value?.removeEventListener('pointercancel', releaseCarouselPointer);
  window.removeEventListener('resize', handleResize);
  clearTimeout(resizeTimer);
  master?.kill();
  ScrollTrigger.getAll().forEach((trigger) => trigger.kill());
});
</script>

<template>
  <div class="stage-showcase">
    <section class="spacer">
      <div class="spacer__inner">
        <div class="spacer__copy">
          <h1 class="spacer__title">
            {{ props.content.title }}
          </h1>
          <p class="spacer__subtitle">
            {{ props.content.subtitle }}
          </p>
        </div>

        <div class="spacer__skyline" aria-hidden="true">
          <img
            src="/home-stage/carousel/1613383ojE1fL80qbZphhF.png"
            alt=""
            class="spacer__skyline-image"
          >
        </div>

        <section
          ref="carouselRoot"
          class="home-carousel"
          :aria-label="props.content.carouselAriaLabel"
        >
          <div class="home-carousel__viewport">
            <div class="home-carousel__track">
              <article
                v-for="slide in carouselSlides"
                :key="slide.key"
                class="home-carousel__slide"
              >
                <img
                  :src="slide.imageUrl"
                  :alt="slide.imageAlt"
                  class="home-carousel__image"
                >
              </article>
            </div>
          </div>
        </section>
      </div>
    </section>

    <section class="stage">
      <div class="layers">
        <svg
          v-for="(module, index) in modules"
          :key="module.code"
          class="layer"
          viewBox="0 0 100 100"
          preserveAspectRatio="none"
        >
          <defs>
            <mask
              :id="`mask${index + 1}`"
              maskUnits="userSpaceOnUse"
            >
              <rect
                x="0"
                y="0"
                width="100"
                height="100"
                fill="black"
              />
              <g :id="`blinds${index + 1}`" />
            </mask>
          </defs>

          <image
            :href="module.stageImagePath"
            x="0"
            y="0"
            width="100"
            height="100"
            preserveAspectRatio="xMidYMid slice"
            :mask="`url(#mask${index + 1})`"
          />
        </svg>

        <div class="stage__shade" />

        <div class="progress-bar">
          <div
            v-for="(_, index) in modules"
            :key="`segment-${index}`"
            class="segment"
          >
            <div class="fill" />
          </div>
        </div>

        <div class="texts">
          <div
            v-for="module in modules"
            :key="`text-${module.code}`"
            class="txt"
          >
            <h1>{{ module.displayTitle }}</h1>
            <h2>{{ module.availabilityLabel }}</h2>
            <span>{{ module.description }}</span>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<style>
.stage-showcase {
  position: relative;
  background: rgb(var(--color-canvas));
}

html[data-theme='dark-neutral'] .stage-showcase {
  background: rgb(var(--color-canvas));
}

html[data-theme='default'] .stage-showcase {
  background: rgb(var(--color-canvas));
}

.spacer {
  min-height: calc(100svh + 5.1rem);
  position: relative;
  overflow: clip;
}

.spacer::before {
  display: none;
  position: absolute;
  inset: 8vh 10vw auto;
  height: 24rem;
  border-radius: 999px;
  background: rgb(var(--color-primary-soft));
  filter: blur(18px);
  content: '';
  pointer-events: none;
}

.spacer__inner {
  position: relative;
  z-index: 1;
  box-sizing: border-box;
  display: grid;
  width: 100%;
  min-height: calc(100svh + 5.1rem);
  padding-top: clamp(5.8rem, 10vh, 7.5rem);
  padding-right: clamp(1rem, 3vw, 2.5rem);
  padding-left: clamp(1rem, 3vw, 2.5rem);
  padding-bottom: clamp(0.8rem, 1.5vh, 1.4rem);
  gap: clamp(1.2rem, 2.4vh, 1.9rem);
  grid-template-rows: minmax(0, 1fr) auto;
}

.spacer__copy {
  display: grid;
  align-content: center;
  justify-items: center;
  gap: 0.9rem;
  width: 100%;
  text-align: center;
  transform: translateY(-16%);
}

.spacer__title {
  margin: 0;
  font-family: var(--font-display);
  font-size: clamp(2.7rem, 5.5vw, 4.8rem);
  line-height: 0.98;
  letter-spacing: 0;
  color: rgb(var(--color-text));
}

.spacer__subtitle {
  margin: 0;
  max-width: 44rem;
  font-size: clamp(0.94rem, 1.3vw, 1.08rem);
  line-height: 1.8;
  color: rgb(var(--color-text-muted));
}

.spacer__skyline {
  position: absolute;
  left: 50%;
  bottom: clamp(13.5rem, 24vh, 18rem);
  width: 100vw;
  z-index: 0;
  pointer-events: none;
  transform: translateX(-50%);
}

.spacer__skyline-image {
  display: block;
  width: 100%;
  max-width: none;
  height: auto;
  max-height: clamp(24rem, 50vh, 38rem);
  margin-inline: auto;
  mix-blend-mode: screen;
  opacity: 0.76;
  filter:
    drop-shadow(0 18px 30px rgb(var(--color-primary) / 0.06))
    saturate(0.82)
    contrast(1.02);
}

.home-carousel {
  position: relative;
  z-index: 1;
  width: 100%;
  height: clamp(17rem, 38svh, 25.5rem);
  margin-bottom: 24px;
  display: grid;
  place-items: center;
  overflow: hidden;
  perspective: 1800px;
  overscroll-behavior: none;
  transform: translateY(-5%);
  transform-style: preserve-3d;
  -webkit-user-select: none;
  user-select: none;
}

.home-carousel__viewport {
  position: relative;
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  overflow: hidden;
  padding-block: clamp(0.9rem, 2.2vh, 1.5rem);
}

.home-carousel.carousel-mode {
  touch-action: none;
  cursor: grab;
}

.home-carousel.carousel-mode.dragging {
  cursor: grabbing;
}

.home-carousel__track {
  position: relative;
  width: 100%;
  height: 100%;
  transform-style: preserve-3d;
}

.home-carousel__slide {
  position: absolute;
  top: 50%;
  left: 50%;
  display: block;
  width: min(24vw, 340px);
  aspect-ratio: 16 / 10;
  overflow: hidden;
  border: 1px solid rgb(255 255 255 / 0.48);
  border-radius: 3px;
  background: rgb(var(--color-surface) / 0.08);
  box-shadow: 0 14px 30px rgb(22 24 34 / 0.16);
  transform-origin: 90% center;
  backface-visibility: hidden;
  contain: layout paint;
  will-change: transform, filter, opacity;
  transition:
    opacity 0.24s ease,
    filter 0.24s ease;
}

.home-carousel__image {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: center center;
  pointer-events: none;
  -webkit-user-drag: none;
  -webkit-user-select: none;
  user-select: none;
}

.stage {
  height: 440vh;
  margin-top: -10vh;
  position: relative;
  z-index: 2;
}

.layers {
  position: sticky;
  top: 0;
  width: 100%;
  height: 100vh;
  overflow: hidden;
  background: rgb(var(--color-surface-muted));
  box-shadow:
    0 -18px 40px rgb(var(--color-text) / 0.08),
    var(--shadow-floating);
}

.layer {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
}

.layer image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  filter: brightness(0.82);
}

.stage__shade {
  position: absolute;
  inset: 0;
  background:
    linear-gradient(180deg, rgb(7 9 16 / 0.08) 0%, rgb(7 9 16 / 0.06) 26%, rgb(7 9 16 / 0.5) 100%);
  z-index: 2;
  pointer-events: none;
}

.texts {
  position: absolute;
  inset: 0;
  z-index: 3;
  padding: 3vw;
  pointer-events: auto;
  -webkit-user-select: text;
  user-select: text;
}

.txt {
  position: absolute;
  inset: 0;
  display: grid;
  align-content: end;
  gap: 1rem;
  padding: clamp(1.5rem, 4vw, 3rem);
  color: rgb(255 255 255 / 0.96);
  text-transform: uppercase;
  clip-path: inset(100% 0 0 0);
  transform: translateY(40px);
  text-shadow: 0 12px 34px rgb(0 0 0 / 0.24);
}

.txt h1 {
  margin: 0;
  max-width: 11ch;
  font-size: clamp(3rem, 6vw, 7.2rem);
  letter-spacing: 0;
  line-height: 0.94;
}

.txt h2 {
  margin: 0;
  font-size: clamp(0.78rem, 0.95vw, 0.96rem);
  letter-spacing: 0.18em;
  color: rgb(255 255 255 / 0.78);
}

.txt span {
  display: block;
  max-width: 30rem;
  font-size: clamp(0.88rem, 1.05vw, 1rem);
  line-height: 1.78;
  text-transform: none;
  color: rgb(255 255 255 / 0.88);
}

.progress-bar {
  position: absolute;
  left: 50%;
  bottom: 2rem;
  z-index: 4;
  width: min(30rem, 72vw);
  display: grid;
  gap: 0.6rem;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  transform: translateX(-50%);
}

.segment {
  height: 0.2rem;
  overflow: hidden;
  border-radius: 999px;
  background: rgb(255 255 255 / 0.18);
}

.fill {
  width: 0;
  height: 100%;
  border-radius: inherit;
  background: rgb(255 255 255 / 0.96);
}

@media (max-width: 1023px) {
  .home-carousel__slide {
    width: min(30rem, 44vw);
  }
}

@media (max-width: 767px) {
  .spacer {
    min-height: calc(100svh + 4.8rem);
  }

  .spacer__inner {
    min-height: calc(100svh + 4.8rem);
    padding-top: 5.25rem;
    padding-bottom: 0.8rem;
  }

  .spacer__title {
    font-size: clamp(2.35rem, 10vw, 3.45rem);
  }

  .spacer__copy {
    transform: translateY(-10%);
  }

  .spacer__skyline {
    bottom: clamp(11.5rem, 20vh, 15rem);
  }

  .spacer__skyline-image {
    width: 100%;
    max-height: clamp(16rem, 34vh, 24rem);
  }

  .home-carousel {
    height: clamp(15.5rem, 35svh, 20.5rem);
  }

  .home-carousel__slide {
    width: min(22rem, 72vw);
  }

  .stage {
    height: 380vh;
  }

  .txt {
    padding: 1.2rem;
  }

  .txt h1 {
    font-size: clamp(2.55rem, 11vw, 4.2rem);
  }

  .txt span {
    max-width: 100%;
  }

  .progress-bar {
    bottom: 1.25rem;
    width: min(22rem, calc(100vw - 2rem));
  }
}
</style>
