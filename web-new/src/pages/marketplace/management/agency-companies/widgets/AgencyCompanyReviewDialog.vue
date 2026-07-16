<!--
 * 管理端 - 代理資料審核詳情彈窗。
 * 1. 按個人代理或代理公司展示完整欄位與審核文件。
 * 2. 提供通過及拒絕操作，拒絕時要求填寫原因。
-->
<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

import type { AgencyAsset, StaffAgencyProfile } from '@/model/agency-profile';
import AppIcon from '@/shared/components/base/AppIcon.vue';
import { usePreferenceStore } from '@/stores/preferences';
import { formatDate } from '@/utils/format';

const props = defineProps<{ company: StaffAgencyProfile; isReviewing: boolean; reviewNote: string }>();
const emit = defineEmits<{ approve: []; close: []; reject: []; 'update:reviewNote': [value: string] }>();
const { t } = useI18n();
const preferenceStore = usePreferenceStore();

const statusLabel = computed(() => t(`marketplace.management.agencyCompany.status.${props.company.status}`));
const profileIcon = computed<'building' | 'user'>(() => props.company.profile_type === 'company' ? 'building' : 'user');
const submittedAt = computed(() => props.company.submitted_at
  ? formatDate(props.company.submitted_at, preferenceStore.locale) : '-');
const formatPhone = (code?: string, number?: string): string => [code, number].filter(Boolean).join(' ') || '-';
const updateReviewNote = (event: Event): void => emit('update:reviewNote', (event.target as HTMLTextAreaElement).value);

// 1. 只在 Staff 審核彈窗展示證明文件
const reviewAssets = computed<Array<{ label: string; asset: AgencyAsset }>>(() => {
  const values = [
    [t('account.agencyCompany.avatarCustom'), props.company.avatar_asset],
    [t('account.agencyCompany.wechatQrImage'), props.company.wechat_qr_asset],
    [t('account.agencyCompany.logo'), props.company.logo_asset],
    [t('account.agencyCompany.eaaLicense'), props.company.eaa_license_asset],
    [t('account.agencyCompany.businessRegistration'), props.company.business_registration_asset],
    [t('account.agencyCompany.companyCard'), props.company.company_card_asset],
  ] as Array<[string, AgencyAsset | null | undefined]>;
  return values.filter((item): item is [string, AgencyAsset] => Boolean(item[1])).map(([label, asset]) => ({ label, asset }));
});
</script>

<template>
  <Teleport to="body">
    <Transition name="management-dialog" appear>
      <div
        class="management-dialog"
        role="dialog"
        aria-modal="true"
        :aria-label="company.name_zh || company.name_en"
        @click.self="emit('close')"
      >
        <section class="management-dialog-panel agency-review-dialog">
          <header class="management-dialog-header agency-review-header">
            <div class="agency-review-identity">
              <span class="agency-review-icon">
                <AppIcon :name="profileIcon" :size="20" />
              </span>
              <div>
                <p>{{ company.profile_type === 'individual' ? t('marketplace.management.agencyCompany.individualType') : t('marketplace.management.agencyCompany.companyType') }}</p>
                <h2>{{ company.name_zh || company.name_en }}</h2>
                <span>{{ company.name_en || company.owner.display_name || '-' }}</span>
              </div>
            </div>
            <div class="agency-review-header-actions">
              <span
                class="management-status-pill agency-review-status"
                :class="`agency-review-status--${company.status}`"
              >
                {{ statusLabel }}
              </span>
              <button
                type="button"
                class="management-dialog-close"
                :aria-label="t('marketplace.management.closeDialog')"
                :disabled="isReviewing"
                @click="emit('close')"
              >
                <AppIcon name="close" :size="18" />
              </button>
            </div>
          </header>

          <div class="management-dialog-body agency-review-body">
            <div class="agency-review-meta">
              <div>
                <span>{{ t('marketplace.management.agencyCompany.submittedAt', { date: submittedAt }) }}</span>
              </div>
              <div>
                <span>{{ t('marketplace.management.agencyCompany.licenseNumber') }}</span>
                <strong>{{ company.license_number || '-' }}</strong>
              </div>
            </div>

            <section class="agency-review-section">
              <header><AppIcon name="user" :size="17" /><h3>{{ t('marketplace.management.agencyCompany.ownerDetails') }}</h3></header>
              <dl class="detail-grid">
                <div><dt>{{ t('marketplace.management.agencyCompany.owner') }}</dt><dd>{{ company.owner.display_name || '-' }}</dd></div>
                <div><dt>{{ t('marketplace.management.agencyCompany.ownerEmail') }}</dt><dd>{{ company.owner.email || '-' }}</dd></div>
                <div><dt>{{ t('marketplace.management.agencyCompany.ownerId') }}</dt><dd>{{ company.owner.public_id || '-' }}</dd></div>
              </dl>
            </section>

            <section class="agency-review-section">
              <header><AppIcon :name="profileIcon" :size="17" /><h3>{{ company.profile_type === 'individual' ? t('account.agencyCompany.personalDetails') : t('account.agencyCompany.companyDetails') }}</h3></header>
              <dl class="detail-grid">
                <div><dt>{{ t('account.agencyCompany.nameZh') }}</dt><dd>{{ company.name_zh || '-' }}</dd></div>
                <div><dt>{{ t('account.agencyCompany.nameEn') }}</dt><dd>{{ company.name_en || '-' }}</dd></div>
                <template v-if="company.profile_type === 'company'">
                  <div class="wide"><dt>{{ t('account.agencyCompany.addressZh') }}</dt><dd>{{ company.address_zh || '-' }}</dd></div>
                  <div class="wide"><dt>{{ t('account.agencyCompany.addressEn') }}</dt><dd>{{ company.address_en || '-' }}</dd></div>
                  <div><dt>{{ t('account.agencyCompany.bigFour') }}</dt><dd>{{ company.is_big_four ? t('account.agencyCompany.yes') : t('account.agencyCompany.no') }}</dd></div>
                </template>
                <template v-else>
                  <div><dt>{{ t('account.agencyCompany.overseasAgent') }}</dt><dd>{{ company.is_overseas ? t('account.agencyCompany.yes') : t('account.agencyCompany.no') }}</dd></div>
                  <div><dt>{{ t('account.agencyCompany.defaultAvatar') }}</dt><dd>{{ company.default_avatar || '-' }}</dd></div>
                  <div><dt>{{ t('account.agencyCompany.wechatId') }}</dt><dd>{{ company.wechat_id || '-' }}</dd></div>
                  <div><dt>{{ t('account.agencyCompany.wechatQrUrl') }}</dt><dd>{{ company.wechat_url || '-' }}</dd></div>
                  <div class="wide"><dt>{{ t('account.agencyCompany.signatureZh') }}</dt><dd>{{ company.signature_zh || '-' }}</dd></div>
                  <div class="wide"><dt>{{ t('account.agencyCompany.signatureEn') }}</dt><dd>{{ company.signature_en || '-' }}</dd></div>
                </template>
                <div><dt>{{ t('marketplace.management.agencyCompany.phoneOne') }}</dt><dd>{{ formatPhone(company.phone_1_country_code, company.phone_1_number) }}<small>WhatsApp · {{ company.phone_1_whatsapp ? t('account.agencyCompany.yes') : t('account.agencyCompany.no') }}</small></dd></div>
                <div><dt>{{ t('marketplace.management.agencyCompany.phoneTwo') }}</dt><dd>{{ formatPhone(company.phone_2_country_code, company.phone_2_number) }}<small>WhatsApp · {{ company.phone_2_whatsapp ? t('account.agencyCompany.yes') : t('account.agencyCompany.no') }}</small></dd></div>
              </dl>
            </section>

            <section class="agency-review-section">
              <header><AppIcon name="picture" :size="17" /><h3>{{ t('marketplace.management.agencyCompany.documents') }}</h3></header>
              <div class="asset-grid">
                <a
                  v-for="item in reviewAssets"
                  :key="item.asset.media_asset_id"
                  class="review-document"
                  :href="item.asset.url"
                  target="_blank"
                  rel="noreferrer"
                >
                  <img :src="item.asset.url" :alt="item.label">
                  <span>{{ item.label }}</span>
                  <small><AppIcon name="view" :size="14" />{{ t('marketplace.management.agencyCompany.reviewAction') }}</small>
                </a>
                <p v-if="reviewAssets.length === 0" class="agency-review-empty">{{ t('marketplace.management.agencyCompany.noDocuments') }}</p>
              </div>
            </section>

            <label v-if="company.status === 'pending'" class="management-dialog-field agency-review-note">
              <span>{{ t('marketplace.management.agencyCompany.reviewNote') }}</span>
              <textarea
                :value="reviewNote"
                :placeholder="t('marketplace.management.agencyCompany.reviewNotePlaceholder')"
                :disabled="isReviewing"
                @input="updateReviewNote"
              ></textarea>
            </label>
            <p v-if="company.status === 'rejected' && company.review_note" class="previous-note"><strong>{{ t('marketplace.management.agencyCompany.previousReviewNote') }}</strong> {{ company.review_note }}</p>
          </div>

          <footer class="management-dialog-actions agency-review-actions">
            <button type="button" class="management-dialog-button" :disabled="isReviewing" @click="emit('close')">{{ t('marketplace.management.cancelAction') }}</button>
            <template v-if="company.status === 'pending'">
              <button type="button" class="management-dialog-button reject" :disabled="isReviewing" @click="emit('reject')"><AppIcon name="close" :size="15" />{{ t('marketplace.management.agencyCompany.rejectAction') }}</button>
              <button type="button" class="management-dialog-button management-dialog-button--primary" :disabled="isReviewing" @click="emit('approve')"><AppIcon name="check-circle" :size="15" />{{ isReviewing ? t('marketplace.management.agencyCompany.reviewing') : t('marketplace.management.agencyCompany.approveAction') }}</button>
            </template>
          </footer>
        </section>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.agency-review-dialog {
  display: grid;
  grid-template-rows: auto minmax(0, 1fr) auto;
  width: min(100%, 58rem);
  max-height: calc(100svh - 2rem);
  overflow: hidden;
}

.agency-review-header {
  align-items: center;
  padding: 1rem 1.1rem;
}

.agency-review-identity,
.agency-review-header-actions,
.agency-review-section > header,
.review-document small {
  display: flex;
  align-items: center;
}

.agency-review-identity {
  min-width: 0;
  gap: 0.75rem;
}

.agency-review-icon {
  display: grid;
  width: 2.65rem;
  height: 2.65rem;
  flex: 0 0 auto;
  place-items: center;
  border: 1px solid rgb(var(--color-border));
  border-radius: 6px;
  background: rgb(var(--color-surface-muted));
  color: rgb(var(--color-primary));
}

.agency-review-identity > div {
  min-width: 0;
}

.agency-review-identity h2,
.agency-review-identity span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.agency-review-identity span {
  display: block;
  margin-top: 0.2rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.76rem;
}

.agency-review-header-actions {
  flex: 0 0 auto;
  gap: 0.55rem;
}

.agency-review-status--pending {
  background: rgb(var(--color-warning-bg));
  color: rgb(var(--color-warning));
}

.agency-review-status--approved {
  background: rgb(var(--color-success-bg));
  color: rgb(var(--color-success));
}

.agency-review-status--rejected {
  background: rgb(var(--color-danger-bg));
  color: rgb(var(--color-danger));
}

.agency-review-body {
  min-height: 0;
  overflow-y: auto;
  gap: 0;
  padding: 0;
}

.agency-review-meta {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(12rem, auto);
  gap: 1rem;
  border-bottom: 1px solid rgb(var(--color-border));
  background: rgb(var(--color-surface-muted));
  padding: 0.75rem 1.1rem;
}

.agency-review-meta div {
  display: grid;
  gap: 0.18rem;
}

.agency-review-meta span,
.detail-grid dt {
  color: rgb(var(--color-text-muted));
  font-size: 0.7rem;
  font-weight: 750;
}

.agency-review-meta strong {
  color: rgb(var(--color-text));
  font-size: 0.82rem;
}

.agency-review-section {
  display: grid;
  gap: 0.85rem;
  border-bottom: 1px solid rgb(var(--color-border));
  padding: 1rem 1.1rem;
}

.agency-review-section > header {
  gap: 0.45rem;
  color: rgb(var(--color-primary));
}

.agency-review-section h3 {
  margin: 0;
  color: rgb(var(--color-text));
  font-size: 0.88rem;
  font-weight: 780;
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.85rem 1rem;
  margin: 0;
}

.detail-grid .wide {
  grid-column: 1 / -1;
}

.detail-grid div {
  min-width: 0;
}

.detail-grid dd {
  margin: 0.22rem 0 0;
  overflow-wrap: anywhere;
  color: rgb(var(--color-text));
  font-size: 0.82rem;
  font-weight: 650;
  line-height: 1.5;
}

.detail-grid dd small {
  display: block;
  margin-top: 0.15rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.7rem;
  font-weight: 600;
}

.asset-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(8rem, 1fr));
  gap: 0.7rem;
}

.review-document {
  display: grid;
  grid-template-rows: auto auto auto;
  overflow: hidden;
  border: 1px solid rgb(var(--color-border));
  border-radius: 6px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
  text-decoration: none;
  transition: border-color 0.18s ease, background 0.18s ease;
}

.review-document:hover {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-surface-muted));
}

.review-document img {
  width: 100%;
  aspect-ratio: 4 / 3;
  object-fit: contain;
  background: rgb(var(--color-surface-muted));
}

.review-document > span {
  overflow: hidden;
  padding: 0.5rem 0.55rem 0.2rem;
  font-size: 0.72rem;
  font-weight: 750;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.review-document small {
  gap: 0.3rem;
  padding: 0.15rem 0.55rem 0.5rem;
  color: rgb(var(--color-primary));
  font-size: 0.68rem;
  font-weight: 700;
}

.agency-review-empty {
  grid-column: 1 / -1;
  margin: 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.8rem;
}

.agency-review-note {
  margin: 1rem 1.1rem;
}

.agency-review-note textarea {
  min-height: 6rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 4px;
  background: rgb(var(--color-surface));
  padding: 0.7rem 0.8rem;
  color: rgb(var(--color-text));
  font: inherit;
  line-height: 1.5;
  outline: none;
  resize: vertical;
}

.agency-review-note textarea:focus {
  border-color: rgb(var(--color-primary));
  box-shadow: 0 0 0 2px rgb(var(--color-primary) / 0.12);
}

.previous-note {
  margin: 1rem 1.1rem;
  border-left: 3px solid rgb(var(--color-danger));
  background: rgb(var(--color-danger-bg));
  padding: 0.7rem 0.8rem;
  color: rgb(var(--color-danger));
  font-size: 0.78rem;
  line-height: 1.55;
}

.agency-review-actions {
  background: rgb(var(--color-surface-raised));
}

.agency-review-actions .management-dialog-button {
  gap: 0.4rem;
}

.reject {
  border-color: rgb(var(--color-danger));
  color: rgb(var(--color-danger));
}

@media (max-width: 767px) {
  .agency-review-dialog {
    width: 100%;
    max-height: calc(100svh - var(--app-safe-top) - var(--app-safe-bottom) - 12px);
  }

  .agency-review-header {
    align-items: flex-start;
  }

  .agency-review-status {
    display: none;
  }

  .agency-review-meta,
  .detail-grid {
    grid-template-columns: 1fr;
  }

  .detail-grid .wide {
    grid-column: auto;
  }

  .asset-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
