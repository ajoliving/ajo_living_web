<!--
 * 首頁物業資訊。
 * 1. 以三張本地題材封面展示經核對的原始報道及公告。
 * 2. 不抓取外站圖片或內容；封面使用既有庫存圖，不視作現場照片。
-->
<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

const { t, locale } = useI18n();
const articles = [
  {
    key: 'access',
    href: 'https://www.hk01.com/article/60400435',
    date: '2026-10-20',
    image: '/home_stage/community_cover.webp',
  },
  {
    key: 'meeting',
    href: 'https://www.pmsa.org.hk/tc/press-release/20260814',
    englishHref: 'https://www.pmsa.org.hk/en/press-release/20260814',
    date: '2026-08-14',
    image: '/home_stage/property_sale.webp',
  },
  {
    key: 'lift',
    href: 'https://www.stheadline.com/society/3627484',
    date: '2026-10-18',
    image: '/home_stage/serviced_apartment.webp',
  },
] as const;

const dayFormatter = computed(() => new Intl.DateTimeFormat(locale.value, {
  dateStyle: 'medium',
  timeZone: 'Asia/Hong_Kong',
}));

// 1. 以香港日期展示原始來源的刊登時間。
const formatDate = (date: string): string => dayFormatter.value.format(new Date(`${date}T00:00:00+08:00`));
</script>

<template>
  <section class="home-property-news" aria-labelledby="home-property-news-title">
    <div class="home-section-heading">
      <h2 id="home-property-news-title">{{ t('home.community.news.title') }}</h2>
    </div>
    <div class="home-news-grid">
      <a
        v-for="article in articles"
        :key="article.key"
        class="home-news-card"
        :href="locale === 'en' && 'englishHref' in article ? article.englishHref : article.href"
      >
        <div class="home-news-cover" aria-hidden="true">
          <img :src="article.image" alt="" loading="lazy" decoding="async" />
          <div class="home-news-cover-copy">
            <span>{{ t(`home.community.news.${article.key}.cover`) }}</span>
            <strong>{{ t(`home.community.news.${article.key}.emphasis`) }}</strong>
          </div>
        </div>
        <div class="home-news-body">
          <h3>{{ t(`home.community.news.${article.key}.title`) }}</h3>
          <p>
            <time :datetime="article.date">{{ formatDate(article.date) }}</time>
            <span>{{ t(`home.community.news.${article.key}.source`) }}</span>
          </p>
        </div>
      </a>
    </div>
  </section>
</template>

<style scoped lang="scss">
/* 1. 內容卡等寬，封面上下分區保留可讀的標題。 */
.home-news-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16px;
}

.home-news-card {
  display: flex;
  flex-direction: column;
  min-width: 0;
  overflow: hidden;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  color: rgb(var(--color-text));
  background: rgb(var(--color-surface));
}

.home-news-card:hover {
  border-color: rgb(var(--color-primary));
}

.home-news-cover {
  display: grid;
  grid-template-rows: 1fr auto;
  aspect-ratio: 16 / 9;
  overflow: hidden;
}

.home-news-cover img {
  width: 100%;
  height: 100%;
  min-height: 0;
  object-fit: cover;
}

.home-news-cover-copy {
  display: grid;
  align-content: center;
  gap: 3px;
  min-height: 75px;
  padding: 10px 14px;
  background: rgb(var(--color-surface-muted));
  font-size: 17px;
  font-weight: 650;
  line-height: 1.25;
}

.home-news-cover-copy strong {
  font-size: 13px;
  font-weight: 650;
  color: rgb(var(--color-text-muted));
}

.home-news-body {
  display: flex;
  flex: 1;
  flex-direction: column;
  justify-content: space-between;
  gap: 8px;
  padding: 12px 14px;
}

.home-news-body h3 {
  margin: 0;
  overflow-wrap: anywhere;
  font-size: 15px;
  font-weight: 650;
  line-height: 1.45;
}

.home-news-body p {
  display: flex;
  flex-wrap: wrap;
  gap: 2px 10px;
  margin: 0;
  color: rgb(var(--color-text-muted));
  font-size: 12px;
}

@media (max-width: 767px) {
  .home-news-grid {
    grid-auto-columns: minmax(208px, 76%);
    grid-auto-flow: column;
    grid-template-columns: none;
    gap: 10px;
    overflow-x: auto;
    padding: 3px 4px 10px;
    margin-inline: -4px;
    scroll-snap-type: x proximity;
  }

  .home-news-card {
    scroll-snap-align: start;
  }
}
</style>
