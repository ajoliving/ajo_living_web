<!--
 * 服務住宅頁。
 * 1. 承接服務住宅主路由入口。
 * 2. 以正式訊息提示目前頻道狀態。
-->
<script setup lang="ts">
import { RouterLink } from 'vue-router';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();
</script>

<template>
  <main class="channel-page">
    <section class="channel-page__hero">
      <div>
        <p class="channel-page__kicker">
          {{ t('nav.channelPreparing') }}
        </p>
        <h1>{{ t('nav.servicedResidences') }}</h1>
        <p>{{ t('nav.servicedResidencesDescription') }}</p>
      </div>
      <RouterLink
        to="/serviced-residences"
        class="channel-page__link channel-page__link--primary"
      >
        {{ t('servicedResidence.channel.openList') }}
      </RouterLink>
    </section>

    <section class="channel-page__grid">
      <article class="channel-page__card">
        <span>{{ t('servicedResidence.channel.public') }}</span>
        <strong>{{ t('servicedResidence.channel.publicTitle') }}</strong>
        <p>{{ t('servicedResidence.channel.publicDescription') }}</p>
      </article>
      <article class="channel-page__card">
        <span>{{ t('servicedResidence.channel.operation') }}</span>
        <strong>{{ t('servicedResidence.channel.operationTitle') }}</strong>
        <p>{{ t('servicedResidence.channel.operationDescription') }}</p>
      </article>
      <article class="channel-page__card">
        <span>{{ t('servicedResidence.channel.next') }}</span>
        <strong>{{ t('servicedResidence.channel.nextTitle') }}</strong>
        <p>{{ t('servicedResidence.channel.nextDescription') }}</p>
      </article>
    </section>
  </main>
</template>

<style scoped>
.channel-page {
  display: grid;
  width: 100%;
  max-width: var(--layout-page-max-width);
  gap: 16px;
  margin: 0 auto;
  padding: 24px var(--layout-page-padding-inline) 72px;
  color: rgb(var(--color-text));
}

.channel-page__hero {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 16px;
  border-bottom: 1px solid rgb(var(--color-border));
  padding-bottom: 18px;
}

.channel-page__kicker {
  margin: 0;
  color: rgb(var(--color-text-muted));
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.18em;
  line-height: 1;
  text-transform: uppercase;
}

.channel-page__hero h1 {
  margin: 10px 0 0;
  font-family: var(--font-display);
  font-size: 36px;
  font-weight: 500;
  line-height: 1.14;
}

.channel-page__hero p:last-child {
  max-width: 560px;
  margin: 10px 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  line-height: 1.8;
}

.channel-page__link {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 38px;
  padding: 0 14px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  font-size: 12px;
  font-weight: 600;
  text-decoration: none;
}

.channel-page__link--primary {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
}

.channel-page__grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.channel-page__card {
  display: grid;
  gap: 8px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface));
  padding: 16px;
}

.channel-page__card span {
  color: rgb(var(--color-text-muted));
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.16em;
  line-height: 1;
  text-transform: uppercase;
}

.channel-page__card strong {
  color: rgb(var(--color-text));
  font-size: 14px;
  font-weight: 600;
  line-height: 1.45;
}

.channel-page__card p {
  margin: 0;
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  line-height: 1.7;
}

@media (max-width: 767px) {
  .channel-page {
    padding: 18px var(--layout-page-padding-inline) 96px;
  }

  .channel-page__hero,
  .channel-page__grid {
    grid-template-columns: 1fr;
    flex-direction: column;
    align-items: start;
  }

  .channel-page__hero h1 {
    font-size: 28px;
  }

  .channel-page__grid {
    display: grid;
  }
}
</style>
