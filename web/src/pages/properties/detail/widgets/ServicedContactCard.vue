<!--
 * 服務式住宅聯絡卡片。
 * 1. 顯示立即預訂主操作、網站與微信入口。
 * 2. 承接聯絡方式解鎖後的電話與微信展示。
-->
<script setup lang="ts">
import { useI18n } from 'vue-i18n';

import AppIcon from '@/shared/components/base/AppIcon.vue';

interface ServicedContactCardProps {
  primaryWhatsAppHref: string;
  websiteUrl: string;
  weChatValue: string;
  phoneValue: string;
  hasContactAccess: boolean;
  loadingContact: boolean;
}

const props = defineProps<ServicedContactCardProps>();

const emit = defineEmits<{
  revealContact: [];
}>();

const { t } = useI18n();
</script>

<template>
  <section class="property-serviced-side-card property-serviced-summary-card">
    <div class="property-serviced-card-header">
      <p class="property-serviced-side-label">{{ t('property.serviced.bookingTitle') }}</p>
    </div>
    <div class="property-serviced-action-stack">
      <a
        v-if="props.primaryWhatsAppHref"
        class="property-serviced-action property-serviced-action--primary"
        :href="props.primaryWhatsAppHref"
        target="_blank"
        rel="noopener noreferrer"
      >
        <AppIcon name="message" :size="18" />
        <span>{{ t('property.editor.whatsappInquiryAction') }}</span>
      </a>
      <button
        v-else
        type="button"
        class="property-serviced-action property-serviced-action--primary"
        @click="emit('revealContact')"
      >
        <AppIcon name="message" :size="18" />
        <span>{{ props.loadingContact ? t('property.detail.loadingContact') : t('property.editor.whatsappInquiryAction') }}</span>
      </button>
      <div class="property-serviced-secondary-actions">
        <a
          v-if="props.weChatValue"
          class="property-serviced-action property-serviced-action--secondary-icon"
          :href="`mailto:?subject=WeChat%3A${props.weChatValue}`"
          :title="`WeChat: ${props.weChatValue}`"
        >
          <AppIcon name="message" :size="18" />
        </a>
        <a
          v-if="props.websiteUrl"
          class="property-serviced-action property-serviced-action--secondary-icon"
          :href="props.websiteUrl"
          target="_blank"
          rel="noopener noreferrer"
          :title="t('property.editor.visitWebsiteAction')"
        >
          <AppIcon name="globe" :size="18" />
        </a>
      </div>
    </div>
    <div
      v-if="!props.hasContactAccess || props.phoneValue || props.weChatValue"
      class="property-serviced-visible-contact"
    >
      <button
        v-if="!props.hasContactAccess && !props.phoneValue && !props.weChatValue"
        type="button"
        class="property-serviced-reveal-button"
        :disabled="props.loadingContact"
        @click="emit('revealContact')"
      >
        <AppIcon name="phone" :size="16" />
        <span>{{ props.loadingContact ? t('property.detail.loadingContact') : t('property.detail.revealContact') }}</span>
      </button>
      <div
        v-if="props.phoneValue"
        class="property-serviced-contact-info"
      >
        <AppIcon name="phone" :size="16" />
        <span>{{ props.phoneValue }}</span>
      </div>
      <div
        v-if="props.weChatValue"
        class="property-serviced-contact-info property-serviced-contact-info--wechat"
      >
        <AppIcon name="message" :size="16" />
        <span>{{ props.weChatValue }}</span>
      </div>
    </div>
  </section>
</template>

<style scoped lang="scss">
.property-serviced-side-card {
  overflow: hidden;
  border: 1px solid rgb(var(--serviced-theme-border));
  border-radius: 0.5rem;
  background: #fff;
  box-shadow: 0 2px 8px rgb(0 0 0 / 0.06);
}

.property-serviced-card-header {
  border-bottom: 1px solid rgb(var(--serviced-theme-border));
  background: rgb(var(--serviced-theme-muted));
  padding: 0.95rem 1.1rem;
}

.property-serviced-action-stack {
  display: grid;
  gap: 0.65rem;
  margin-top: 1rem;
}

.property-serviced-secondary-actions {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0;
}

.property-serviced-action {
  display: inline-flex;
  width: 100%;
  min-height: 2.8rem;
  align-items: center;
  justify-content: center;
  gap: 0.6rem;
  border: none;
  border-radius: 0;
  background: transparent;
  color: #fff;
  font-size: 0.95rem;
  font-weight: 900;
  text-decoration: none;
  cursor: pointer;
  transition: all 0.2s ease-in-out;
}

.property-serviced-action--primary {
  border-bottom: 1px solid rgb(var(--serviced-theme-border));
  background: #25d366;
  color: #fff;
}

.property-serviced-action--primary:hover {
  background: #1cb14f;
}

.property-serviced-action--secondary-icon {
  border: 1px solid rgb(var(--serviced-theme-border));
  border-left: none;
  background: #fff;
  color: rgb(var(--serviced-theme-text));
}

.property-serviced-action--secondary-icon:hover {
  background: rgb(var(--serviced-theme-muted));
}

.property-serviced-action--secondary-icon:first-child {
  border-left: 1px solid rgb(var(--serviced-theme-border));
}

.property-serviced-visible-contact {
  display: grid;
  gap: 0;
  margin: 0;
  border-top: 1px solid rgb(var(--serviced-theme-border));
  padding: 0;
}

.property-serviced-reveal-button {
  display: flex;
  min-height: 2.6rem;
  align-items: center;
  justify-content: center;
  gap: 0.6rem;
  border: none;
  border-bottom: 1px solid rgb(var(--serviced-theme-border));
  background: rgb(var(--serviced-theme-muted));
  color: rgb(var(--serviced-theme-primary));
  font-size: 0.9rem;
  font-weight: 900;
  cursor: pointer;
  transition: all 0.2s ease-in-out;
  padding: 0 1rem;
}

.property-serviced-reveal-button:hover {
  background: rgb(244 250 248);
}

.property-serviced-reveal-button:disabled {
  opacity: 0.6;
}

.property-serviced-contact-info {
  display: flex;
  min-height: 2.6rem;
  align-items: center;
  justify-content: center;
  gap: 0.8rem;
  padding: 0 1rem;
  border-bottom: 1px solid rgb(var(--serviced-theme-border));
  background: #fff;
  color: rgb(var(--serviced-theme-text));
  font-size: 0.9rem;
  font-weight: 900;
  text-decoration: none;
  transition: all 0.2s ease-in-out;
}

.property-serviced-contact-info:last-child {
  border-bottom: none;
}

.property-serviced-contact-info:hover {
  background: rgb(var(--serviced-theme-muted));
}

.property-serviced-contact-info--wechat {
  cursor: default;
}

.property-serviced-contact-info--wechat:hover {
  background: #fff;
}

.property-serviced-contact-info .app-icon {
  flex: 0 0 auto;
  color: rgb(var(--serviced-theme-primary));
}

.property-serviced-contact-info span {
  min-width: 0;
  overflow: hidden;
  text-align: center;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
