<!--
 * 我的物業發布頁。
 * 1. 根據頻道讀取我的樓盤或服務式住宅列表。
 * 2. 支援發布、重發、成交與下架等狀態操作。
 * 3. 在會員中心內以彈窗承載新增與編輯發布流程。
-->
<script setup lang="ts">
import axios from 'axios';
import { computed, onMounted, ref } from 'vue';
import { RouterLink } from 'vue-router';
import { useI18n } from 'vue-i18n';

import {
  getPropertyOptionLabel,
  propertyAdPackageOptions,
  propertyTransactionTypeOptions,
  propertyTypeOptions,
  servicedAdPackageOptions,
} from '@/constants/property';
import {
  deactivatePropertySale,
  deactivateServicedApartment,
  fetchMyPropertySaleListings,
  fetchMyServicedApartmentListings,
  markPropertySaleSold,
  publishPropertySale,
  publishServicedApartment,
  republishPropertySale,
  republishServicedApartment,
  renewPropertySale,
} from '@/httpapis/properties';
import type { PaginationMeta } from '@/model/api';
import type {
  PropertyChannel,
  PropertyListParams,
  PropertyListingSummaryResponse,
} from '@/model/property';
import AppActionConfirmDialog from '@/shared/components/base/AppActionConfirmDialog.vue';
import AppIcon from '@/shared/components/base/AppIcon.vue';
import { useDialogBackdropClose } from '@/shared/composables/useDialogBackdropClose';
import { useFeedbackStore } from '@/stores/feedback';
import { usePreferenceStore } from '@/stores/preferences';
import { useSessionStore } from '@/stores/session';
import { formatDate } from '@/utils/format';
import {
  resolvePropertyArea,
  resolvePropertyCommunityName,
  resolvePropertyCoverImage,
  resolvePropertyDetailPath,
  resolvePropertyDistrict,
  resolvePropertyPriceText,
  resolvePropertyRooms,
  resolvePropertyStatus,
  resolvePropertySummary,
  resolvePropertyTitle,
} from '@/utils/property';
import { formatAjoPoints, resolveWalletChargeCost } from '@/utils/wallet';

import PropertyEditorPage from '../editor/PropertyEditorPage.vue';

const props = defineProps<{
  channel: PropertyChannel;
  basePath?: string;
}>();

type MyPropertyTab = 'all' | 'draft' | 'active' | 'hidden' | 'expired' | 'sold';
type MyPropertyAction = 'publish' | 'republish' | 'renew' | 'mark-sold' | 'deactivate';
interface PropertyEditorDialogStepPayload {
  activeIndex: number;
  total: number;
}

type PropertyEditorDialogInstance = InstanceType<typeof PropertyEditorPage> & {
  requestCloseEditor: () => Promise<void>;
};

const { t } = useI18n();
const feedbackStore = useFeedbackStore();
const preferenceStore = usePreferenceStore();
const sessionStore = useSessionStore();

const loading = ref(false);
const activeTab = ref<MyPropertyTab>('all');
const searchQuery = ref('');
const items = ref<PropertyListingSummaryResponse[]>([]);
const pagination = ref<PaginationMeta>({
  page: 1,
  page_size: 10,
  total: 0,
});
const isEditorOpen = ref(false);
const editorListingId = ref('');
const editorStepIndex = ref(0);
const editorStepTotal = ref(4);
const propertyEditorDialog = ref<PropertyEditorDialogInstance | null>(null);
const pendingAction = ref<{ action: MyPropertyAction; listing: PropertyListingSummaryResponse } | null>(null);
const actionLoading = ref(false);

const pageTitle = computed(() =>
  props.channel === 'sale' ? t('property.sale.myTitle') : t('property.serviced.myTitle'),
);
const currentBasePath = computed(() =>
  props.basePath || (
    props.channel === 'sale' ? '/properties/my' : '/serviced-residences/my'
  ),
);
const chargeCost = computed(() =>
  resolveWalletChargeCost(props.channel === 'sale' ? 'property_sale' : 'serviced_apartment'),
);
const editorDialogTitle = computed(() => {
  if (editorListingId.value) {
    return t('property.mine.editListingTitle');
  }

  return props.channel === 'sale'
    ? t('property.sale.publishTitle')
    : t('property.serviced.publishTitle');
});
const editorDialogModeLabel = computed(() =>
  editorListingId.value ? t('property.editor.editMode') : t('property.editor.publishMode'),
);
const editorDialogSteps = computed(() => [
  {
    key: 'category',
    label: props.channel === 'sale' ? t('property.editor.stepCategory') : t('property.editor.servicedStepCategory'),
  },
  {
    key: 'ad',
    label: t('property.editor.adPackageField'),
  },
  {
    key: 'details',
    label: props.channel === 'sale' ? t('property.editor.saleTitle') : t('property.editor.servicedTitle'),
  },
  {
    key: 'contact',
    label: t('property.editor.contact'),
  },
]);
const editorDialogKey = computed(() =>
  `${props.channel}-${editorListingId.value || 'new'}-${isEditorOpen.value ? 'open' : 'closed'}`,
);
const formatPoints = (value: number): string =>
  formatAjoPoints(value, t('common.brand.pointsName'), preferenceStore.locale);
const resolveActionCharge = (action: MyPropertyAction, listing: PropertyListingSummaryResponse): number => {
  if (action === 'publish') {
    const publishTotal = listing.publish_points_total ?? chargeCost.value;
    const draftPointsPaid = listing.draft_points_paid ?? 0;
    return Math.max(publishTotal - draftPointsPaid, 0);
  }
  if (action === 'republish') {
    return listing.property_sale?.ad_price_points ?? listing.serviced_apartment?.ad_price_points ?? chargeCost.value;
  }
  if (action === 'renew') {
    const publishCost = listing.property_sale?.ad_price_points ?? chargeCost.value;
    return Math.ceil(publishCost / 2);
  }
  return 0;
};
const resolvePublishActionLabel = (listing: PropertyListingSummaryResponse): string => {
  const pointsDue = resolveActionCharge('publish', listing);
  if (pointsDue === 0) {
    return t('property.mine.publish');
  }
  return `${t('property.mine.publish')} · ${formatPoints(pointsDue)}`;
};
const tabOptions = computed<Array<{ label: string; value: MyPropertyTab }>>(() => [
  { label: t('property.mine.all'), value: 'all' },
  { label: t('property.mine.draft'), value: 'draft' },
  { label: t('property.mine.active'), value: 'active' },
  { label: t('property.mine.hidden'), value: 'hidden' },
  { label: t('property.mine.expired'), value: 'expired' },
  { label: t('property.mine.sold'), value: 'sold' },
]);
const filteredItems = computed(() => {
  return items.value;
});
const hasPrevious = computed(() => pagination.value.page > 1);
const hasNext = computed(() => pagination.value.page * pagination.value.page_size < pagination.value.total);
const paginationText = computed(() => {
  if (pagination.value.total <= 0) {
    return t('property.mine.paginationEmpty');
  }

  const start = (pagination.value.page - 1) * pagination.value.page_size + 1;
  const end = Math.min(pagination.value.page * pagination.value.page_size, pagination.value.total);

  return t('property.mine.paginationRange', {
    start,
    end,
    total: pagination.value.total,
  });
});
const confirmDescription = computed(() => {
  const action = pendingAction.value?.action;
  if (!action) {
    return '';
  }

  const listing = pendingAction.value?.listing;
  if (!listing) {
    return '';
  }
  if (action === 'publish') {
    const due = resolveActionCharge(action, listing);
    const paid = listing.draft_points_paid ?? 0;
    const total = listing.publish_points_total ?? chargeCost.value;
    if (due === 0 && paid >= total) {
      return t('property.mine.confirmPublishDraftFullyPaid');
    }
    if (paid > 0) {
      return t('property.mine.confirmPublishDraftCharge', {
        due: formatPoints(due),
        total: formatPoints(total),
        paid: formatPoints(paid),
      });
    }
    return `${t('property.mine.confirmPublishCharge')} ${formatPoints(due)}`;
  }
  if (action === 'republish') {
    return `${t('property.mine.confirmRepublishCharge')} ${formatPoints(resolveActionCharge(action, listing))}`;
  }
  if (action === 'renew') {
    return `${t('property.mine.confirmRenewCharge')} ${formatPoints(resolveActionCharge(action, listing))}`;
  }

  if (action === 'deactivate') {
    return t('property.mine.confirmDeactivate');
  }

  return t('property.sale.soldConfirm');
});
const confirmActionLabel = computed(() => {
  const action = pendingAction.value?.action;
  if (action === 'mark-sold') {
    return t('property.sale.soldAction');
  }

  return action ? t(`property.mine.${action}`) : '';
});

// 1.1 重置彈窗步驟狀態
const resetEditorStepState = (): void => {
  editorStepIndex.value = 0;
  editorStepTotal.value = 4;
};

// 1. 讀取我的發布
const loadMyListings = async (targetPage = pagination.value.page): Promise<void> => {
  loading.value = true;

  try {
    const params: PropertyListParams = {
      page: targetPage,
      page_size: pagination.value.page_size,
      keyword: searchQuery.value.trim() || undefined,
      status: activeTab.value === 'all' ? '' : activeTab.value,
    };
    const response = props.channel === 'sale'
      ? await fetchMyPropertySaleListings(params)
      : await fetchMyServicedApartmentListings(params);

    items.value = response.data.data.items;
    pagination.value = response.data.data.pagination;
  } catch (error: unknown) {
    feedbackStore.pushToast(
      axios.isAxiosError<{ message?: string }>(error)
        ? error.response?.data?.message ?? t('property.mine.loadError')
        : t('property.mine.loadError'),
      'error',
    );
    items.value = [];
    pagination.value = {
      ...pagination.value,
      page: targetPage,
      total: 0,
    };
  } finally {
    loading.value = false;
  }
};

// 2. 切換列表狀態
const setActiveTab = async (tab: MyPropertyTab): Promise<void> => {
  if (activeTab.value === tab) {
    return;
  }

  activeTab.value = tab;
  await loadMyListings(1);
};

// 3. 搜尋我的發布
const searchListings = async (): Promise<void> => {
  await loadMyListings(1);
};

// 4. 讀取上一頁
const loadPreviousPage = async (): Promise<void> => {
  if (!hasPrevious.value || loading.value) {
    return;
  }

  await loadMyListings(pagination.value.page - 1);
};

// 5. 讀取下一頁
const loadNextPage = async (): Promise<void> => {
  if (!hasNext.value || loading.value) {
    return;
  }

  await loadMyListings(pagination.value.page + 1);
};

// 6. 開啟新增發布彈窗
const openCreateEditor = (): void => {
  resetEditorStepState();
  editorListingId.value = '';
  isEditorOpen.value = true;
};

// 7. 開啟編輯發布彈窗
const openEditEditor = (listingId: string): void => {
  resetEditorStepState();
  editorListingId.value = listingId;
  isEditorOpen.value = true;
};

// 8. 關閉發布彈窗
const closeEditor = (): void => {
  isEditorOpen.value = false;
  editorListingId.value = '';
  resetEditorStepState();
};

// 9. 請求關閉發布彈窗
const requestCloseEditor = (): void => {
  void propertyEditorDialog.value?.requestCloseEditor();
};

// 9.1 只在完整點擊遮罩時請求關閉編輯器
const {
  handleBackdropPointerCancel: handleEditorBackdropPointerCancel,
  handleBackdropPointerDown: handleEditorBackdropPointerDown,
  handleBackdropPointerUp: handleEditorBackdropPointerUp,
} = useDialogBackdropClose(requestCloseEditor);

// 9.2 同步彈窗步驟狀態
const handleEditorStepChange = (payload: PropertyEditorDialogStepPayload): void => {
  editorStepIndex.value = payload.activeIndex;
  editorStepTotal.value = payload.total;
};

// 10. 處理草稿儲存完成
const handleEditorSaved = async (): Promise<void> => {
  const isNewListing = editorListingId.value.trim() === '';
  closeEditor();
  if (isNewListing) {
    activeTab.value = 'draft';
    await loadMyListings(1);
    return;
  }

  await loadMyListings(pagination.value.page);
};

// 11. 處理發布完成
const handleEditorPublished = async (): Promise<void> => {
  closeEditor();
  activeTab.value = 'active';
  await loadMyListings(1);
};

// 12. 輸出可顯示日期
const formatOptionalDate = (value?: string | null): string =>
  value ? formatDate(value, preferenceStore.locale) : '-';

// 13. 輸出表格價格
const resolvePriceText = (listing: PropertyListingSummaryResponse): string =>
  resolvePropertyPriceText(listing, preferenceStore.locale);

// 14. 輸出地區及位置
const resolveLocationText = (listing: PropertyListingSummaryResponse): string =>
  `${resolvePropertyCommunityName(listing, preferenceStore.locale)} · ${resolvePropertyDistrict(listing, preferenceStore.locale)}`;

// 15. 輸出面積及房型
const resolveSpecText = (listing: PropertyListingSummaryResponse): string => {
  const area = resolvePropertyArea(listing);
  const areaText = area > 0 ? t('common.unit.sqft', { value: area }) : '-';
  return `${areaText} · ${resolvePropertyRooms(listing, preferenceStore.locale)}`;
};

// 16. 輸出類型
const resolveTypeText = (listing: PropertyListingSummaryResponse): string => {
  if (listing.property_sale) {
    const transaction = propertyTransactionTypeOptions.find(
      (option) => option.value === listing.property_sale?.transaction_type,
    );
    const propertyType = propertyTypeOptions.find(
      (option) => option.value === listing.property_sale?.property_type,
    );

    return [
      transaction ? getPropertyOptionLabel(transaction, preferenceStore.locale) : '',
      propertyType ? getPropertyOptionLabel(propertyType, preferenceStore.locale) : '',
    ].filter(Boolean).join(' · ') || '-';
  }

  return props.channel === 'serviced'
    ? t('property.serviced.title')
    : '-';
};

// 17. 輸出推廣方案
const resolveAdPackageText = (listing: PropertyListingSummaryResponse): string => {
  const packageCode =
    listing.property_sale?.ad_package_code ||
    listing.serviced_apartment?.ad_package_code ||
    'basic';
  const options = listing.serviced_apartment ? servicedAdPackageOptions : propertyAdPackageOptions;
  const matched = options.find((option) => option.value === packageCode);

  return matched ? getPropertyOptionLabel(matched, preferenceStore.locale) : packageCode;
};

// 18. 開啟狀態操作確認
const requestAction = (action: MyPropertyAction, listing: PropertyListingSummaryResponse): void => {
  if (!actionLoading.value) {
    pendingAction.value = { action, listing };
  }
};

// 19. 執行已確認的狀態操作
const confirmAction = async (): Promise<void> => {
  const target = pendingAction.value;
  if (!target || actionLoading.value) {
    return;
  }

  pendingAction.value = null;
  actionLoading.value = true;
  try {
    const listingId = target.listing.listing_id;
    if (props.channel === 'sale') {
      if (target.action === 'publish') {
        await publishPropertySale(listingId);
      }
      if (target.action === 'republish') {
        await republishPropertySale(listingId);
      }
      if (target.action === 'renew') {
        await renewPropertySale(listingId);
      }
      if (target.action === 'mark-sold') {
        await markPropertySaleSold(listingId);
      }
      if (target.action === 'deactivate') {
        await deactivatePropertySale(listingId);
      }
    } else {
      if (target.action === 'publish') {
        await publishServicedApartment(listingId);
      }
      if (target.action === 'republish') {
        await republishServicedApartment(listingId);
      }
      if (target.action === 'deactivate') {
        await deactivateServicedApartment(listingId);
      }
    }

    feedbackStore.pushToast(t('property.mine.statusUpdated'), 'success');
    if (target.action === 'publish' || target.action === 'republish' || target.action === 'renew') {
      await sessionStore.loadCurrentUser();
    }
    if (target.action === 'publish' || target.action === 'republish') {
      activeTab.value = 'active';
      await loadMyListings(1);
      return;
    }

    await loadMyListings(pagination.value.page);
  } catch (error: unknown) {
    feedbackStore.pushToast(
      axios.isAxiosError<{ message?: string }>(error)
        ? error.response?.data?.message ?? t('property.mine.updateError')
        : t('property.mine.updateError'),
      'error',
    );
  } finally {
    actionLoading.value = false;
  }
};

// 20. 關閉狀態操作確認
const cancelAction = (): void => {
  pendingAction.value = null;
};

// 21. 輸出狀態顯示
const resolveStatusLabel = (listing: PropertyListingSummaryResponse): string =>
  t(`property.mine.${resolvePropertyStatus(listing)}`);

onMounted(() => {
  void loadMyListings(1);
});
</script>

<template>
  <main class="property-my-page">
    <section class="property-my-heading">
      <div>
        <p class="property-kicker">
          {{ props.channel === 'sale' ? t('property.sale.title') : t('property.serviced.title') }}
        </p>
        <h1>{{ pageTitle }}</h1>
      </div>
      <button
        type="button"
        class="property-button property-button--primary"
        @click="openCreateEditor"
      >
        <AppIcon
          name="plus-square"
          :size="17"
        />
        {{ t('property.list.publish') }}
      </button>
    </section>

    <form
      class="property-my-toolbar"
      @submit.prevent="searchListings"
    >
      <div class="property-search-row">
        <label class="property-search-input">
          <AppIcon
            name="search"
            class="member-search-input-icon"
            :size="17"
          />
          <input
            v-model="searchQuery"
            type="search"
            :placeholder="t('property.list.keywordPlaceholder')"
          />
        </label>
        <button
          type="submit"
          class="property-button property-button--secondary"
          :disabled="loading"
        >
          <AppIcon
            name="search"
            :size="16"
          />
          {{ t('property.mine.search') }}
        </button>
      </div>
      <div class="property-tab-row">
        <button
          v-for="tab in tabOptions"
          :key="tab.value"
          type="button"
          class="property-tab"
          :class="activeTab === tab.value ? 'property-tab--active' : ''"
          @click="setActiveTab(tab.value)"
        >
          {{ tab.label }}
        </button>
      </div>
    </form>

    <section
      v-if="loading"
      class="property-empty"
    >
      {{ t('common.status.loading') }}
    </section>

    <section
      v-else-if="filteredItems.length === 0"
      class="property-empty"
    >
      <p>{{ t('property.list.noListings') }}</p>
      <button
        type="button"
        class="property-button property-button--primary"
        @click="openCreateEditor"
      >
        <AppIcon
          name="plus-square"
          :size="17"
        />
        {{ t('property.list.publish') }}
      </button>
    </section>

    <section
      v-else
      class="property-panel"
    >
      <div class="property-panel__title">
        <h2>{{ pageTitle }}</h2>
        <span>{{ paginationText }}</span>
      </div>

      <div class="property-manage-list property-table-wrap">
        <article
          v-for="listing in filteredItems"
          :key="listing.listing_id"
          class="property-manage-card"
        >
          <div class="property-manage-card__overview">
            <div class="property-item__media">
              <img
                v-if="resolvePropertyCoverImage(listing)"
                :src="resolvePropertyCoverImage(listing)?.url"
                :alt="resolvePropertyTitle(listing, preferenceStore.locale)"
              />
              <div
                v-else
                class="property-item__placeholder"
              >
                <AppIcon
                  name="picture"
                  :size="22"
                />
              </div>
            </div>

            <div class="property-item__content">
              <div class="property-item__heading">
                <strong>{{ resolvePropertyTitle(listing, preferenceStore.locale) }}</strong>
                <span
                  class="property-status-pill"
                  :class="`property-status-pill--${resolvePropertyStatus(listing)}`"
                >
                  {{ resolveStatusLabel(listing) }}
                </span>
              </div>
              <p>{{ resolvePropertySummary(listing, preferenceStore.locale) }}</p>
              <strong class="property-item__price">{{ resolvePriceText(listing) }}</strong>
            </div>
          </div>

          <dl class="property-manage-card__details">
            <div>
              <dt>{{ t('property.mine.columnType') }}</dt>
              <dd>{{ resolveTypeText(listing) }}</dd>
            </div>
            <div>
              <dt>{{ t('property.mine.columnLocation') }}</dt>
              <dd>{{ resolveLocationText(listing) }}</dd>
            </div>
            <div>
              <dt>{{ t('property.mine.columnSpec') }}</dt>
              <dd>{{ resolveSpecText(listing) }}</dd>
            </div>
            <div>
              <dt>{{ t('property.mine.columnAdPackage') }}</dt>
              <dd>{{ resolveAdPackageText(listing) }}</dd>
            </div>
            <div>
              <dt>{{ t('property.mine.columnPublishedAt') }}</dt>
              <dd>{{ formatOptionalDate(listing.published_at) }}</dd>
            </div>
            <div>
              <dt>{{ t('property.mine.columnExpiresAt') }}</dt>
              <dd>{{ formatOptionalDate(listing.expire_at) }}</dd>
            </div>
            <div>
              <dt>{{ t('property.mine.columnUpdatedAt') }}</dt>
              <dd>{{ formatOptionalDate(listing.updated_at) }}</dd>
            </div>
          </dl>

          <div class="property-table-actions">
            <button
              type="button"
              class="property-button property-button--primary"
              @click="openEditEditor(listing.listing_id)"
            >
              {{ t('property.mine.edit') }}
            </button>
            <RouterLink
              :to="resolvePropertyDetailPath(listing)"
              class="property-button property-button--secondary"
            >
              {{ t('property.mine.publicDetail') }}
            </RouterLink>
            <button
              v-if="listing.publication_status === 'draft'"
              type="button"
              class="property-button property-button--primary"
              @click="requestAction('publish', listing)"
            >
              {{ resolvePublishActionLabel(listing) }}
            </button>
            <button
              v-if="listing.publication_status === 'expired' || listing.publication_status === 'hidden'"
              type="button"
              class="property-button property-button--primary"
              @click="requestAction('republish', listing)"
            >
              {{ t('property.mine.republish') }} · {{ formatPoints(resolveActionCharge('republish', listing)) }}
            </button>
            <button
              v-if="channel === 'sale' && listing.publication_status === 'active' && listing.business_status !== 'sold'"
              type="button"
              class="property-button property-button--secondary"
              @click="requestAction('renew', listing)"
            >
              {{ t('property.mine.renew') }} · {{ formatPoints(resolveActionCharge('renew', listing)) }}
            </button>
            <button
              v-if="channel === 'sale' && listing.publication_status === 'active' && listing.business_status !== 'sold'"
              type="button"
              class="property-button property-button--secondary"
              @click="requestAction('mark-sold', listing)"
            >
              {{ t('property.sale.soldAction') }}
            </button>
            <button
              v-if="listing.publication_status === 'active'"
              type="button"
              class="property-button property-button--secondary"
              @click="requestAction('deactivate', listing)"
            >
              {{ t('property.mine.deactivate') }}
            </button>
          </div>
        </article>
      </div>

      <div class="property-pagination">
        <p>{{ paginationText }}</p>
        <div>
          <button
            type="button"
            class="property-button property-button--secondary"
            :disabled="!hasPrevious || loading"
            @click="loadPreviousPage"
          >
            {{ t('property.mine.previous') }}
          </button>
          <button
            type="button"
            class="property-button property-button--secondary"
            :disabled="!hasNext || loading"
            @click="loadNextPage"
          >
            {{ t('property.mine.next') }}
          </button>
        </div>
      </div>
    </section>

    <Teleport to="body">
      <Transition name="property-editor-dialog">
        <div
          v-if="isEditorOpen"
          class="property-editor-dialog"
          role="dialog"
          aria-modal="true"
          :aria-label="editorDialogTitle"
          @pointercancel="handleEditorBackdropPointerCancel"
          @pointerdown="handleEditorBackdropPointerDown"
          @pointerup="handleEditorBackdropPointerUp"
        >
          <section class="property-editor-dialog__panel">
            <header class="property-editor-dialog__header">
              <span class="property-editor-dialog__title">
                {{ editorDialogModeLabel }}
              </span>
              <div
                class="property-editor-dialog__progress"
                :aria-label="t('property.editor.publishSteps')"
              >
                <span
                  v-for="(step, stepIndex) in editorDialogSteps.slice(0, editorStepTotal)"
                  :key="step.key"
                  class="property-editor-dialog__progress-step"
                  :class="{
                    'property-editor-dialog__progress-step--active': stepIndex === editorStepIndex,
                    'property-editor-dialog__progress-step--done': stepIndex < editorStepIndex,
                  }"
                >
                  <span class="property-editor-dialog__progress-number">{{ stepIndex + 1 }}</span>
                  <span class="property-editor-dialog__progress-label">{{ step.label }}</span>
                </span>
              </div>
              <button
                type="button"
                class="property-editor-dialog__close"
                :aria-label="t('common.action.close')"
                @click="requestCloseEditor"
              >
                <AppIcon
                  name="close"
                  :size="18"
                />
              </button>
            </header>

            <div class="property-editor-dialog__body">
              <PropertyEditorPage
                ref="propertyEditorDialog"
                :key="editorDialogKey"
                :channel="channel"
                :listing-id="editorListingId"
                :return-path="currentBasePath"
                embedded
                hide-header
                hide-progress
                @cancel="closeEditor"
                @saved="handleEditorSaved"
                @published="handleEditorPublished"
                @step-change="handleEditorStepChange"
              />
            </div>
          </section>
        </div>
      </Transition>
    </Teleport>

    <AppActionConfirmDialog
      :open="Boolean(pendingAction)"
      :title="t('property.mine.confirmActionTitle')"
      :description="confirmDescription"
      :cancel-label="t('property.mine.cancelAction')"
      :confirm-label="confirmActionLabel"
      :confirming="actionLoading"
      @cancel="cancelAction"
      @confirm="confirmAction"
    />
  </main>
</template>

<style scoped>
.property-my-page {
  display: grid;
  gap: 1rem;
  width: 100%;
  color: rgb(var(--color-text));
}

.property-my-heading,
.property-my-toolbar {
  display: grid;
  gap: 1rem;
}

.property-my-heading {
  align-items: end;
  margin-bottom: 0;
}

.property-kicker {
  margin: 0 0 0.6rem;
  color: rgb(var(--color-primary));
  font-size: 0.75rem;
  font-weight: 900;
  letter-spacing: 0;
  text-transform: uppercase;
}

.property-my-heading h1 {
  margin: 0;
  font-family: var(--font-display);
  font-size: 2rem;
  font-weight: 500;
  line-height: 1.2;
}

.property-button {
  display: inline-flex;
  min-height: 2.6rem;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  border-radius: 0.5rem;
  padding: 0 0.9rem;
  font-size: 0.86rem;
  font-weight: 900;
  cursor: pointer;
}

.property-button:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.property-button--primary {
  border: 1px solid rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
}

.property-button--secondary {
  border: 1px solid rgb(var(--color-border));
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
}

.property-my-toolbar {
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.75rem;
  background: rgb(var(--color-surface));
  padding: 1rem;
}

.property-search-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 0.75rem;
}

.property-search-input {
  display: flex;
  min-height: 2.75rem;
  align-items: center;
  gap: 0.5rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.5rem;
  background: rgb(var(--color-surface-raised));
  padding: 0 0.8rem;
}

.property-search-input input {
  width: 100%;
  border: 0;
  background: transparent;
  color: inherit;
  outline: 0;
}

.property-tab-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.property-tab {
  min-height: 2.25rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: var(--radius-sm);
  background: rgb(var(--color-surface-raised));
  padding: 0 0.75rem;
  color: rgb(var(--color-text-muted));
  font-weight: 900;
}

.property-tab--active {
  border-color: rgb(var(--color-primary));
  color: rgb(var(--color-primary));
}

.property-my-stack {
  display: grid;
  gap: 1rem;
  margin-top: 1rem;
}

.property-status-pill {
  flex: 0 0 auto;
  border-radius: 999px;
  background: rgb(var(--color-primary-soft));
  padding: 0.3rem 0.55rem;
  color: rgb(var(--color-primary));
  font-size: 0.75rem;
  font-weight: 900;
}

.property-status-pill--active,
.property-status-pill--published {
  background: rgb(var(--color-success-bg));
  color: rgb(var(--color-success));
}

.property-status-pill--expired {
  background: rgb(var(--color-warning-bg));
  color: rgb(var(--color-warning));
}

.property-status-pill--hidden,
.property-status-pill--draft {
  background: rgb(var(--color-surface-muted));
  color: rgb(var(--color-text-muted));
}

.property-panel {
  display: grid;
  gap: 14px;
  margin-top: 0;
}

.property-panel__title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  border-bottom: 1px solid rgb(var(--color-border));
  padding-bottom: 12px;
}

.property-panel__title h2 {
  margin: 0;
  color: rgb(var(--color-primary));
  font-size: 14px;
  font-weight: 600;
}

.property-panel__title span {
  color: rgb(var(--color-text-muted));
  font-size: 12px;
}

.property-table-wrap {
  min-width: 0;
}

.property-manage-list {
  display: grid;
  gap: 10px;
}

.property-manage-card {
  display: grid;
  gap: 14px;
  min-width: 0;
  border: 1px solid rgb(var(--color-border));
  border-radius: 6px;
  background: rgb(var(--color-surface));
  padding: 14px;
}

.property-manage-card__overview {
  display: grid;
  grid-template-columns: 128px minmax(0, 1fr);
  gap: 14px;
  min-width: 0;
  align-items: start;
}

.property-manage-card__details {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(8.5rem, 1fr));
  gap: 1px;
  overflow: hidden;
  margin: 0;
  border: 1px solid rgb(var(--color-border));
  border-radius: 4px;
  background: rgb(var(--color-surface-raised));
}

.property-manage-card__details > div {
  min-width: 0;
  background: rgb(var(--color-surface));
  padding: 10px 12px;
}

.property-manage-card__details dt {
  margin-bottom: 4px;
  color: rgb(var(--color-text-muted));
  font-size: 10px;
  font-weight: 600;
}

.property-manage-card__details dd {
  overflow: hidden;
  margin: 0;
  color: rgb(var(--color-text));
  font-size: 12px;
  font-weight: 600;
  line-height: 1.45;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.property-item__media {
  overflow: hidden;
  width: 100%;
  border: 1px solid rgb(var(--color-border));
  border-radius: 4px;
  background: rgb(var(--color-surface-muted));
  aspect-ratio: 4 / 3;
}

.property-item__media img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.property-item__placeholder {
  display: grid;
  height: 100%;
  place-items: center;
  color: rgb(var(--color-text-muted));
}

.property-item__content {
  display: grid;
  align-content: start;
  gap: 8px;
  min-width: 0;
}

.property-item__heading {
  display: flex;
  min-width: 0;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
}

.property-item__heading > strong {
  overflow: hidden;
  color: rgb(var(--color-text));
  font-size: 16px;
  font-weight: 700;
  line-height: 1.35;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.property-item__content p {
  display: -webkit-box;
  overflow: hidden;
  margin: 0;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  line-height: 1.6;
}

.property-item__price {
  color: rgb(var(--color-primary));
  font-family: var(--font-display);
  font-size: 18px;
  font-weight: 600;
}

.property-table-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  border-top: 1px solid rgb(var(--color-border));
  padding-top: 12px;
}

.property-table-actions .property-button {
  min-width: 5.5rem;
}

.property-pagination {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  border-top: 1px solid rgb(var(--color-border));
  padding-top: 12px;
}

.property-pagination p {
  margin: 0;
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  line-height: 1.5;
}

.property-pagination div {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 8px;
}

.property-empty {
  display: grid;
  justify-items: center;
  gap: 12px;
  margin-top: 1rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.75rem;
  background: rgb(var(--color-surface));
  padding: 2rem;
  color: rgb(var(--color-text-muted));
  text-align: center;
}

.property-empty p {
  margin: 0;
}

.property-editor-dialog {
  position: fixed;
  z-index: 110;
  inset: 0;
  display: grid;
  place-items: center;
  background: rgb(15 23 42 / 0.34);
  padding: 16px;
  backdrop-filter: blur(2px);
}

.property-editor-dialog__panel {
  display: grid;
  grid-template-rows: auto minmax(0, 1fr);
  width: min(96vw, 1260px);
  height: min(820px, calc(100svh - 32px));
  max-height: calc(100svh - 32px);
  overflow: hidden;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  box-shadow: 0 24px 70px rgb(15 23 42 / 0.24);
  justify-self: center;
}

.property-editor-dialog__header {
  display: grid;
  grid-template-columns: minmax(7rem, auto) minmax(0, 34rem) auto;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  border-bottom: 1px solid rgb(var(--color-border));
  background: rgb(var(--color-surface));
  padding: 14px 16px;
}

.property-editor-dialog__title {
  color: rgb(var(--color-text));
  font-size: 13px;
  font-weight: 700;
  letter-spacing: 0;
  line-height: 1;
  white-space: nowrap;
}

.property-editor-dialog__progress {
  position: relative;
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  align-items: start;
  width: 100%;
  justify-self: center;
  padding: 0;
}

.property-editor-dialog__progress::before {
  position: absolute;
  top: 14px;
  right: 20px;
  left: 20px;
  height: 1px;
  background: rgb(var(--color-border));
  content: "";
}

.property-editor-dialog__progress-step {
  position: relative;
  z-index: 1;
  display: grid;
  min-width: 0;
  justify-items: center;
  gap: 6px;
  color: rgb(var(--color-text-muted));
}

.property-editor-dialog__progress-number {
  display: inline-flex;
  width: 28px;
  height: 28px;
  align-items: center;
  justify-content: center;
  border: 1px solid rgb(var(--color-border));
  border-radius: 999px;
  background: rgb(var(--color-surface));
  font-size: 11px;
  font-weight: 800;
  line-height: 1;
}

.property-editor-dialog__progress-label {
  overflow: hidden;
  max-width: min(7rem, 100%);
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0;
  line-height: 1.25;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.property-editor-dialog__progress-step--active .property-editor-dialog__progress-number,
.property-editor-dialog__progress-step--done .property-editor-dialog__progress-number {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
}

.property-editor-dialog__progress-step--active .property-editor-dialog__progress-number {
  box-shadow: 0 0 0 4px rgb(var(--color-primary) / 0.12);
}

.property-editor-dialog__progress-step--active .property-editor-dialog__progress-label,
.property-editor-dialog__progress-step--done .property-editor-dialog__progress-label {
  color: rgb(var(--color-primary));
}

.property-editor-dialog__close {
  display: inline-flex;
  width: 34px;
  height: 34px;
  align-items: center;
  justify-content: center;
  border: 1px solid transparent;
  border-radius: 6px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text-muted));
  cursor: pointer;
}

.property-editor-dialog__close:hover {
  border-color: rgb(var(--color-border));
  background: rgb(var(--color-surface-muted));
  color: rgb(var(--color-text));
}

.property-editor-dialog__body {
  min-height: 0;
  overflow-y: auto;
  background: rgb(var(--color-surface-muted));
  padding: 16px;
}

.property-editor-dialog__body :deep(.property-editor-page) {
  max-width: none;
  padding: 0;
}

.property-editor-dialog-enter-active,
.property-editor-dialog-leave-active {
  transition: opacity 0.18s ease;
}

.property-editor-dialog-enter-active .property-editor-dialog__panel,
.property-editor-dialog-leave-active .property-editor-dialog__panel {
  transition:
    opacity 0.18s ease,
    transform 0.18s ease;
}

.property-editor-dialog-enter-from,
.property-editor-dialog-leave-to {
  opacity: 0;
}

.property-editor-dialog-enter-from .property-editor-dialog__panel,
.property-editor-dialog-leave-to .property-editor-dialog__panel {
  opacity: 0;
  transform: translateY(8px);
}

@media (min-width: 760px) {
  .property-my-heading {
    grid-template-columns: minmax(0, 1fr) auto;
  }

  .property-my-card {
    grid-template-columns: 16rem minmax(0, 1fr) 14rem;
  }
}

.property-my-page {
  padding: 0 0 72px;
}

.property-my-heading {
  border-bottom: 1px solid rgb(var(--color-border));
  margin-bottom: 0;
  padding-bottom: 12px;
}

.property-kicker {
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0;
}

.property-my-heading h1 {
  font-size: 26px;
  font-weight: 500;
}

.property-button {
  min-height: 34px;
  border-radius: 2px;
  font-size: 12px;
  font-weight: 600;
}

.property-my-toolbar,
.property-empty {
  border-radius: 3px;
  box-shadow: none;
}

.property-my-toolbar {
  gap: 10px;
  padding: 12px;
}

.property-search-input {
  min-height: 34px;
  border-radius: 2px;
  background: rgb(var(--color-surface));
  font-size: 12px;
}

.property-tab {
  min-height: 32px;
  border-radius: var(--radius-sm);
  background: rgb(var(--color-surface));
  font-size: 11px;
  font-weight: 600;
}

.property-tab--active {
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
}

.property-status-pill {
  border-radius: 2px;
  font-size: 10px;
  font-weight: 600;
}

@media (max-width: 767px) {
  .property-my-page {
    padding-bottom: calc(var(--app-mobile-content-bottom) + 16px);
  }

  .property-search-row {
    grid-template-columns: minmax(0, 1fr) auto;
  }

  .property-my-heading {
    grid-template-columns: minmax(0, 1fr) auto;
    align-items: center;
  }

  .property-my-heading .property-button {
    min-width: 44px;
  }

  .property-tab-row {
    flex-wrap: nowrap;
    margin-inline: -12px;
    padding: 0 12px 2px;
    overflow-x: auto;
    scrollbar-width: none;
  }

  .property-tab-row::-webkit-scrollbar {
    display: none;
  }

  .property-tab {
    flex: 0 0 auto;
  }

  .property-manage-card {
    padding: 12px;
  }

  .property-manage-card__overview {
    grid-template-columns: 96px minmax(0, 1fr);
    gap: 12px;
  }

  .property-item__heading {
    display: grid;
    justify-content: stretch;
  }

  .property-status-pill {
    width: fit-content;
  }

  .property-manage-card__details {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .property-table-actions .property-button {
    flex: 1 1 7rem;
  }

  .property-search-row {
    display: grid;
  }

  .property-pagination {
    align-items: stretch;
    flex-direction: column;
  }

  .property-pagination div {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    justify-content: stretch;
  }

  .property-editor-dialog {
    padding:
      var(--app-safe-top)
      var(--layout-page-padding-inline)
      calc(var(--app-safe-bottom) + 10px);
    place-items: end center;
  }

  .property-editor-dialog__panel {
    width: 100%;
    height: calc(100svh - var(--app-safe-top) - var(--app-safe-bottom) - 20px);
    max-height: calc(100svh - var(--app-safe-top) - var(--app-safe-bottom) - 20px);
    border-radius: 8px 8px 0 0;
  }

  .property-editor-dialog__header {
    grid-template-columns: auto minmax(0, 1fr) auto;
    gap: 10px;
    padding: 10px;
  }

  .property-editor-dialog__close {
    width: 44px;
    height: 44px;
  }

  .property-editor-dialog__body {
    padding: 12px;
  }

  .property-editor-dialog__progress-label {
    display: none;
  }
}

@media (max-width: 380px) {
  .property-manage-card__overview,
  .property-manage-card__details {
    grid-template-columns: minmax(0, 1fr);
  }

  .property-item__media {
    max-height: 11rem;
  }
}
</style>
