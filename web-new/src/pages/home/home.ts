/*
 * 首頁內容配置。
 * 1. 根據語系輸出首頁主標與模組入口文案。
 * 2. 保持首頁內容簡潔，避免展示開發說明與原型話術。
 */
import type { HomeCarouselImage, HomeModuleCard } from '@/model/home-content';

// 1. 定義首頁模組代碼
export type HomeModuleCode = 'secondhand' | 'property_sale' | 'serviced_apartment';

// 2. 定義首頁模組色調
export type HomeModuleTone = 'copper' | 'slate' | 'sage';

// 3. 定義首頁模組圖示
export type HomeModuleIcon = 'browse' | 'location' | 'home';

// 4. 定義首頁 grid 區塊尺寸
export type HomeGridTileSize = 'hero' | 'wide' | 'tall' | 'square';

// 5. 定義首頁 grid 區塊色調
export type HomeGridTileTone = 'accent' | 'soft' | 'contrast' | 'neutral';

// 6. 定義首頁 stage 轉場模式
export type HomeStagePattern = 'random' | 'column';

// 7. 定義首頁操作按鈕
export interface HomeModuleAction {
  label: string;
  to: string;
  variant: 'primary' | 'secondary';
}

// 8. 定義首頁指標卡片
export interface HomeModuleMetric {
  label: string;
  value: string;
}

// 9. 定義首頁 scroll grid 區塊
export interface HomeGridTile {
  id: string;
  eyebrow: string;
  title: string;
  value: string;
  size: HomeGridTileSize;
  tone: HomeGridTileTone;
}

// 10. 定義首頁模組資料
export interface HomeModuleDefinition {
  code: HomeModuleCode;
  index: string;
  tone: HomeModuleTone;
  stagePattern: HomeStagePattern;
  stageImagePath: string;
  carouselImagePath: string;
  iconName: HomeModuleIcon;
  kicker: string;
  displayTitle: string;
  title: string;
  description: string;
  availabilityLabel: string;
  isLive: boolean;
  noteLabel: string;
  statusNote: string;
  chips: string[];
  highlights: string[];
  metrics: HomeModuleMetric[];
  primaryAction?: HomeModuleAction;
  secondaryAction?: HomeModuleAction;
  gridEyebrow: string;
  gridDescription: string;
  gridTiles: HomeGridTile[];
}

// 11. 定義首頁輪播顯示圖
export interface HomeCarouselDisplayImage {
  id: string;
  url: string;
  alt: string;
}

// 12. 定義首頁整體文案
export interface HomeLandingContent {
  eyebrow: string;
  title: string;
  subtitle: string;
  carouselAriaLabel: string;
  primaryCta: string;
  secondaryCta: string;
  heroHighlights: string[];
  heroMetrics: HomeModuleMetric[];
  carouselImages: HomeCarouselDisplayImage[];
  modules: HomeModuleDefinition[];
}

type HomeTranslate = (key: string, params?: Record<string, number | string>) => string;

const formatMetricCount = (count: number) => String(Math.max(count, 0)).padStart(2, '0');

const tValue = (
  t: HomeTranslate,
  key: string,
  params?: Record<string, number | string>,
) => (params ? t(key, params) : t(key));

const buildGridTile = (
  t: HomeTranslate,
  moduleKey: HomeModuleCode,
  tileKey: string,
  size: HomeGridTileSize,
  tone: HomeGridTileTone,
  params?: Record<string, number | string>,
): HomeGridTile => {
  const key = `home.landing.modules.${moduleKey}.gridTiles.${tileKey}`;

  return {
    id: tileKey,
    eyebrow: tValue(t, `${key}.eyebrow`, params),
    title: tValue(t, `${key}.title`, params),
    value: tValue(t, `${key}.value`, params),
    size,
    tone,
  };
};

// 1. buildHomeLandingContent 根據語系建立首頁內容
export const buildHomeLandingContent = (
  t: HomeTranslate,
  featuredCount: number,
  channelCount: number,
  carouselImages: HomeCarouselImage[] = [],
  moduleCards: HomeModuleCard[] = [],
): HomeLandingContent => ({
  eyebrow: t('home.eyebrow'),
  title: t('home.title'),
  subtitle: t('home.subtitle'),
  carouselAriaLabel: t('home.carouselAriaLabel'),
  primaryCta: t('home.primaryCta'),
  secondaryCta: t('home.landing.secondaryCta'),
  heroHighlights: [
    t('home.landing.heroHighlights.primary'),
    t('home.landing.heroHighlights.featured'),
    t('home.landing.heroHighlights.actions'),
  ],
  heroMetrics: [
    { value: '03', label: t('home.landing.heroMetrics.channels') },
    { value: formatMetricCount(channelCount), label: t('home.landing.heroMetrics.sections') },
    { value: formatMetricCount(featuredCount), label: t('home.landing.heroMetrics.featured') },
  ],
  carouselImages: buildCarouselDisplayImages(t, carouselImages),
  modules: applyHomeModuleCards([
    {
      code: 'secondhand',
      index: '01',
      tone: 'copper',
      stagePattern: 'random',
      stageImagePath: '/home-stage/secondhand.webp',
      carouselImagePath: '/home-stage/carousel/building.jpeg',
      iconName: 'browse',
      kicker: t('home.landing.modules.secondhand.kicker'),
      displayTitle: t('home.landing.modules.secondhand.displayTitle'),
      title: t('home.landing.modules.secondhand.title'),
      description: t('home.landing.modules.secondhand.description'),
      availabilityLabel: t('home.landing.modules.secondhand.availabilityLabel'),
      isLive: true,
      noteLabel: t('home.landing.modules.secondhand.noteLabel'),
      statusNote: t('home.landing.modules.secondhand.statusNote'),
      chips: [
        t('home.landing.modules.secondhand.chips.listings'),
        t('home.landing.modules.secondhand.chips.filter'),
        t('home.landing.modules.secondhand.chips.chat'),
      ],
      highlights: [
        t('home.landing.modules.secondhand.highlights.featured'),
        t('home.landing.modules.secondhand.highlights.flow'),
        t('home.landing.modules.secondhand.highlights.chat'),
      ],
      metrics: [
        { label: t('home.landing.modules.secondhand.metrics.featured'), value: formatMetricCount(featuredCount) },
        { label: t('home.landing.modules.secondhand.metrics.source'), value: t('home.landing.modules.secondhand.metrics.sourceValue') },
        { label: t('home.landing.modules.secondhand.metrics.status'), value: t('home.landing.modules.secondhand.metrics.statusValue') },
      ],
      primaryAction: {
        label: t('home.landing.modules.secondhand.primaryAction'),
        to: '/marketplace',
        variant: 'primary',
      },
      secondaryAction: {
        label: t('home.landing.modules.secondhand.secondaryAction'),
        to: '/marketplace/filter',
        variant: 'secondary',
      },
      gridEyebrow: t('home.landing.modules.secondhand.gridEyebrow'),
      gridDescription: t('home.landing.modules.secondhand.gridDescription'),
      gridTiles: [
        buildGridTile(t, 'secondhand', 'live', 'hero', 'accent', { count: formatMetricCount(featuredCount) }),
        buildGridTile(t, 'secondhand', 'scope', 'wide', 'soft'),
        buildGridTile(t, 'secondhand', 'list', 'tall', 'contrast'),
        buildGridTile(t, 'secondhand', 'filter', 'square', 'neutral'),
        buildGridTile(t, 'secondhand', 'chat', 'square', 'soft'),
        buildGridTile(t, 'secondhand', 'order', 'wide', 'accent'),
        buildGridTile(t, 'secondhand', 'member', 'square', 'neutral'),
        buildGridTile(t, 'secondhand', 'status', 'square', 'contrast'),
      ],
    },
    {
      code: 'property_sale',
      index: '02',
      tone: 'slate',
      stagePattern: 'column',
      stageImagePath: '/home-stage/property-sale.webp',
      carouselImagePath: '/home-stage/carousel/intercom.png',
      iconName: 'location',
      kicker: t('home.landing.modules.property_sale.kicker'),
      displayTitle: t('home.landing.modules.property_sale.displayTitle'),
      title: t('home.landing.modules.property_sale.title'),
      description: t('home.landing.modules.property_sale.description'),
      availabilityLabel: t('home.landing.modules.property_sale.availabilityLabel'),
      isLive: true,
      noteLabel: t('home.landing.modules.property_sale.noteLabel'),
      statusNote: t('home.landing.modules.property_sale.statusNote'),
      chips: [
        t('home.landing.modules.property_sale.chips.owner'),
        t('home.landing.modules.property_sale.chips.viewing'),
        t('home.landing.modules.property_sale.chips.leads'),
      ],
      highlights: [
        t('home.landing.modules.property_sale.highlights.inventory'),
        t('home.landing.modules.property_sale.highlights.rhythm'),
        t('home.landing.modules.property_sale.highlights.workflow'),
      ],
      metrics: [
        { label: t('home.landing.modules.property_sale.metrics.status'), value: t('home.landing.modules.property_sale.metrics.statusValue') },
        { label: t('home.landing.modules.property_sale.metrics.audience'), value: t('home.landing.modules.property_sale.metrics.audienceValue') },
        { label: t('home.landing.modules.property_sale.metrics.focus'), value: t('home.landing.modules.property_sale.metrics.focusValue') },
      ],
      primaryAction: {
        label: t('home.landing.modules.property_sale.primaryAction'),
        to: '/properties',
        variant: 'primary',
      },
      secondaryAction: {
        label: t('home.landing.modules.property_sale.secondaryAction'),
        to: '/account/properties/sale/new',
        variant: 'secondary',
      },
      gridEyebrow: t('home.landing.modules.property_sale.gridEyebrow'),
      gridDescription: t('home.landing.modules.property_sale.gridDescription'),
      gridTiles: [
        buildGridTile(t, 'property_sale', 'hero', 'hero', 'contrast'),
        buildGridTile(t, 'property_sale', 'plan', 'wide', 'neutral'),
        buildGridTile(t, 'property_sale', 'viewing', 'tall', 'soft'),
        buildGridTile(t, 'property_sale', 'owner', 'square', 'accent'),
        buildGridTile(t, 'property_sale', 'agent', 'square', 'neutral'),
        buildGridTile(t, 'property_sale', 'lead', 'wide', 'soft'),
        buildGridTile(t, 'property_sale', 'photo', 'square', 'accent'),
        buildGridTile(t, 'property_sale', 'trust', 'square', 'contrast'),
      ],
    },
    {
      code: 'serviced_apartment',
      index: '03',
      tone: 'sage',
      stagePattern: 'random',
      stageImagePath: '/home-stage/serviced-apartment.webp',
      carouselImagePath: '/home-stage/carousel/rant.png',
      iconName: 'home',
      kicker: t('home.landing.modules.serviced_apartment.kicker'),
      displayTitle: t('home.landing.modules.serviced_apartment.displayTitle'),
      title: t('home.landing.modules.serviced_apartment.title'),
      description: t('home.landing.modules.serviced_apartment.description'),
      availabilityLabel: t('home.landing.modules.serviced_apartment.availabilityLabel'),
      isLive: true,
      noteLabel: t('home.landing.modules.serviced_apartment.noteLabel'),
      statusNote: t('home.landing.modules.serviced_apartment.statusNote'),
      chips: [
        t('home.landing.modules.serviced_apartment.chips.shortStay'),
        t('home.landing.modules.serviced_apartment.chips.monthly'),
        t('home.landing.modules.serviced_apartment.chips.moveIn'),
      ],
      highlights: [
        t('home.landing.modules.serviced_apartment.highlights.separate'),
        t('home.landing.modules.serviced_apartment.highlights.amenity'),
        t('home.landing.modules.serviced_apartment.highlights.fit'),
      ],
      metrics: [
        { label: t('home.landing.modules.serviced_apartment.metrics.stay'), value: t('home.landing.modules.serviced_apartment.metrics.stayValue') },
        { label: t('home.landing.modules.serviced_apartment.metrics.service'), value: t('home.landing.modules.serviced_apartment.metrics.serviceValue') },
        { label: t('home.landing.modules.serviced_apartment.metrics.status'), value: t('home.landing.modules.serviced_apartment.metrics.statusValue') },
      ],
      primaryAction: {
        label: t('home.landing.modules.serviced_apartment.primaryAction'),
        to: '/serviced-residences',
        variant: 'primary',
      },
      secondaryAction: {
        label: t('home.landing.modules.serviced_apartment.secondaryAction'),
        to: '/account/properties/serviced-residences/new',
        variant: 'secondary',
      },
      gridEyebrow: t('home.landing.modules.serviced_apartment.gridEyebrow'),
      gridDescription: t('home.landing.modules.serviced_apartment.gridDescription'),
      gridTiles: [
        buildGridTile(t, 'serviced_apartment', 'hero', 'hero', 'soft'),
        buildGridTile(t, 'serviced_apartment', 'term', 'wide', 'accent'),
        buildGridTile(t, 'serviced_apartment', 'movein', 'tall', 'contrast'),
        buildGridTile(t, 'serviced_apartment', 'amenity', 'square', 'neutral'),
        buildGridTile(t, 'serviced_apartment', 'ops', 'square', 'soft'),
        buildGridTile(t, 'serviced_apartment', 'checkin', 'wide', 'neutral'),
        buildGridTile(t, 'serviced_apartment', 'district', 'square', 'accent'),
        buildGridTile(t, 'serviced_apartment', 'tone', 'square', 'contrast'),
      ],
    },
  ], moduleCards),
});

// 2. buildCarouselDisplayImages 建立首頁輪播圖片清單
const buildCarouselDisplayImages = (
  t: HomeTranslate,
  carouselImages: HomeCarouselImage[],
): HomeCarouselDisplayImage[] =>
  carouselImages
    .slice()
    .sort((left, right) => left.sort_order - right.sort_order)
    .map((image, index) => ({
      id: image.media_asset_id,
      url: image.url,
      alt: t('home.carouselImageAlt', { index: index + 1 }),
    }));

// 3. applyHomeModuleCards 套用首頁後台三大圖設定
const applyHomeModuleCards = (
  modules: HomeModuleDefinition[],
  moduleCards: HomeModuleCard[],
): HomeModuleDefinition[] => {
  const cardMap = new Map(moduleCards.map((card) => [card.module_code, card]));

  return modules.map((module) => {
    const card = cardMap.get(module.code);
    if (!card) {
      return module;
    }

    return {
      ...module,
      stageImagePath: card.url || module.stageImagePath,
      carouselImagePath: card.url || module.carouselImagePath,
      displayTitle: card.title || module.displayTitle,
      title: card.title || module.title,
      description: card.body || module.description,
      statusNote: card.subtitle || module.statusNote,
    };
  });
};
