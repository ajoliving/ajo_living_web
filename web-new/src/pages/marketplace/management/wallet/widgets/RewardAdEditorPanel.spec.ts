/*
 * 廣告任務編輯面板測試。
 * 1. 驗證 display 廣告欄位顯示。
 * 2. 驗證 display 廣告鎖定圖片媒體類型。
 */
import { reactive } from 'vue';
import { mount } from '@vue/test-utils';
import { describe, expect, it } from 'vitest';

import { emptyRewardAdForm } from '../wallet';

import RewardAdEditorPanel from './RewardAdEditorPanel.vue';

const t = (key: string): string => {
  const messages: Record<string, string> = {
    'marketplace.management.walletAdEditorTitle': '廣告任務',
    'marketplace.management.walletAdEditorDescription': '設定廣告',
    'marketplace.management.walletAdTypeField': '廣告用途',
    'marketplace.management.walletAdTypeReward': '看廣告賺積分',
    'marketplace.management.walletAdTypeDisplayShort': '右側短廣告',
    'marketplace.management.walletAdTypeDisplayLong': '右側長廣告',
    'marketplace.management.walletAdTitleField': '任務標題',
    'marketplace.management.walletAdSummaryField': '任務摘要',
    'marketplace.management.walletAdTitlePlaceholder': '請輸入標題',
    'marketplace.management.walletAdSummaryPlaceholder': '請輸入摘要',
    'marketplace.management.walletAdMediaTypeField': '媒體類型',
    'marketplace.management.walletAdMediaTypeImage': '圖片廣告',
    'marketplace.management.walletAdMediaTypeVideo': '影片廣告',
    'marketplace.management.walletAdMediaUploadField': '上傳媒體',
    'marketplace.management.walletAdDisplayChannelField': '展示頻道',
    'marketplace.management.walletAdDisplayLayoutField': '展示樣式',
    'marketplace.management.walletAdDisplayLayout.image_full': '長版圖片',
    'marketplace.management.walletAdDisplayLayout.image_text': '短版圖片',
    'marketplace.management.walletAdDisplayShortSize': '16:9',
    'marketplace.management.walletAdDisplayLongSize': '9:16',
    'marketplace.management.walletAdSortOrderField': '排序',
    'marketplace.management.walletAdRetentionDaysField': '保留天數',
    'marketplace.management.walletAdActiveField': '啟用任務',
    'marketplace.management.walletAdTargetField': '跳轉 URL',
    'marketplace.management.walletAdTargetPlaceholder': 'https://',
    'marketplace.management.walletAdCreateAction': '建立任務',
    'marketplace.settings.displayAdTextField': '展示內容',
    'marketplace.settings.displayAdChannelPropertySale': '樓盤放售',
    'marketplace.settings.displayAdChannelFurniture': '二手家私',
    'marketplace.settings.displayAdChannelServicedApartment': '服務式住宅',
  };

  return messages[key] ?? key;
};

describe('RewardAdEditorPanel', () => {
  it('shows display ad controls and disables video selection in display mode', () => {
    const adForm = reactive(emptyRewardAdForm());
    adForm.adType = 'display_long';
    adForm.displayChannel = 'property_sale';
    adForm.displayPlacement = 'listing_side';
    adForm.mediaType = 'image';

    const wrapper = mount(RewardAdEditorPanel, {
      props: {
        adForm,
        isEditingAd: false,
        savingAd: false,
        t,
      },
    });

    expect(wrapper.text()).toContain('展示頻道');
    expect(wrapper.text()).toContain('展示樣式');
    expect(wrapper.text()).toContain('樓盤放售');
    expect(wrapper.find('select:disabled').exists()).toBe(true);
  });
});
