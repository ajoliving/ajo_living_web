/*
 * 物業發布編輯頁測試。
 * 1. 驗證完整樓盤資料會依序建立草稿、更新圖片並發布。
 * 2. 驗證未完成資料可以保存草稿但不會發布。
 * 3. 驗證缺少必填欄位時會指出欄位並阻止發布請求。
 * 4. 驗證缺少相片時會返回資料步驟並阻止發布。
 * 5. 驗證住宅物業資料的欄位順序與精簡內容。
 * 6. 驗證住宅細則按已選欄目展開及清除。
 * 7. 驗證四步發布流程按租售類型顯示價格標題與議價基礎文案。
 * 8. 驗證切換放租後不會提交放售專用價格標記。
 * 9. 驗證繁體中文標題與單位介紹可翻譯並分行顯示 English 欄位。
 * 10. 驗證翻譯失敗不會覆蓋現有 English 內容。
 * 11. 驗證電話1與電話2可分別保存區號及 WhatsApp 聯絡設定。
 * 12. 驗證住宅及服務式住宅缺漏欄位會完整列出、標示並聚焦到具體輸入框。
 * 13. 驗證不同樓盤類型沿用一致的大廈資料順序。
 * 14. 驗證已發布樓盤修改後只保存變更，不重複調用草稿發布接口。
 */
import { flushPromises, mount } from '@vue/test-utils';
import { nextTick } from 'vue';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import {
  createPropertySale,
  createServicedApartment,
  publishPropertySale,
  publishServicedApartment,
  translatePropertyContent,
  updatePropertySale,
} from '@/httpapis/properties';

import PropertyEditorPage from './PropertyEditorPage.vue';

const testMocks = vi.hoisted(() => ({
  pushToast: vi.fn(),
  routerPush: vi.fn(),
  loadCurrentUser: vi.fn(),
}));

vi.mock('vue-router', () => ({
  onBeforeRouteLeave: vi.fn(),
  useRoute: () => ({ params: {} }),
  useRouter: () => ({ push: testMocks.routerPush }),
}));

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string, params?: Record<string, unknown>) => {
      const messages: Record<string, string> = {
        'common.status.loading': '載入中',
        'property.editor.actualFloorField': '實際樓層',
        'property.editor.addressEnField': '街名及門牌(英文)',
        'property.editor.addressField': '街名及門牌',
        'property.editor.areaUnverifiedField': '面積資料未核實',
        'property.editor.blockNameField': '座數 / 大廈',
        'property.editor.buildingAgeField': '樓齡',
        'property.editor.commercialBuildingNameField': '工商大廈名稱',
        'property.editor.contactNameEnField': '英文名',
        'property.editor.directionField': '座向',
        'property.editor.estateNameField': '大廈名稱',
        'property.editor.bedroomField': '房間數量',
        'property.editor.grossAreaField': '建築面積',
        'property.editor.kitchenTypeField': '廚房類型',
        'property.editor.locationAreaField': '區域',
        'property.editor.locationDistrictField': '地區',
        'property.editor.locationSubdistrictField': '分區',
        'property.editor.managementFeeField': '管理費',
        'property.editor.media': '相片',
        'property.editor.phone2Field': '電話2',
        'property.editor.phoneField': '電話1',
        'property.editor.phoneWhatsappField': '可 WhatsApp 聯絡',
        'property.editor.publisherIdentityField': '業主或地產代理',
        'property.editor.imageRequired': '發布前至少需要一張相片',
        'property.editor.priceSuffixPlus': '標示的價格為議價基礎',
        'property.editor.publishNow': '儲存並發布',
        'property.editor.publishSuccess': '已發布',
        'property.editor.republishNow': '儲存並重新發布',
        'property.editor.rentModule': '租金',
        'property.editor.residentialModuleA': '物業資料',
        'property.editor.residentialModuleE': '聯絡人',
        'property.editor.residentialFeatureGroupField': '細則欄目',
        'property.editor.residentialTagAppliances': '連電器',
        'property.editor.residentialTagFeature': '特色',
        'property.editor.residentialTagFitout': '有裝修',
        'property.editor.residentialTagFurniture': '連傢俬',
        'property.editor.residentialTagParking': '連車位',
        'property.editor.residentialTagUnitFeature': '單位特色',
        'property.editor.residentialTagView': '有景觀',
        'property.editor.saveDraft': '保存草稿',
        'property.editor.saveChanges': '儲存修改',
        'property.editor.salePriceModule': '售價',
        'property.editor.saleModuleA': '物業資料',
        'property.editor.requiredFieldsDetail': '請先填寫：{fields}',
        'property.editor.roomTypeName': '房型名稱',
        'property.editor.roomTypeRent': '房型租金',
        'property.editor.titleEnField': '放盤標題(英文)',
        'property.editor.translateToEnglish': '翻譯成 English',
        'property.editor.translatingContent': '翻譯中',
        'property.editor.translationSuccess': '已產生 English 內容',
        'property.editor.translationError': '暫時無法翻譯，請稍後再試。',
        'property.editor.translationSourceRequired': '請先填寫放盤標題及單位介紹',
        'property.editor.updateSuccess': '樓盤修改已儲存',
        'property.editor.unitNameField': '單位',
        'property.editor.usableAreaField': '實用面積',
        'property.editor.wechatField': 'Wechat ID',
        'property.editor.validationSummaryTitle': '請檢查以下必填欄位（{count}項）',
        'property.editor.requiredFieldInline': '請填寫「{field}」',
      };

      return (messages[key] ?? key)
        .replace('{fields}', String(params?.fields ?? ''))
        .replace('{field}', String(params?.field ?? ''))
        .replace('{count}', String(params?.count ?? ''));
    },
  }),
}));

vi.mock('@/stores/feedback', () => ({
  useFeedbackStore: () => ({ pushToast: testMocks.pushToast }),
}));

vi.mock('@/stores/preferences', () => ({
  usePreferenceStore: () => ({ locale: 'zh-HK' }),
}));

vi.mock('@/stores/session', () => ({
  useSessionStore: () => ({
    me: {
      ajo_balance: 5000,
      district_code: '',
      email: '',
    },
    loadCurrentUser: testMocks.loadCurrentUser,
  }),
}));

vi.mock('@/httpapis/properties', () => ({
  createPropertySale: vi.fn(),
  createServicedApartment: vi.fn(),
  fetchPropertySaleDetail: vi.fn(),
  fetchServicedApartmentDetail: vi.fn(),
  publishPropertySale: vi.fn(),
  publishServicedApartment: vi.fn(),
  republishPropertySale: vi.fn(),
  republishServicedApartment: vi.fn(),
  searchPropertyAddresses: vi.fn(),
  translatePropertyContent: vi.fn(),
  updatePropertySale: vi.fn(),
  updateServicedApartment: vi.fn(),
}));

vi.mock('@/httpapis/staff', () => ({
  fetchStaffPropertySaleDetail: vi.fn(),
  fetchStaffServicedApartmentDetail: vi.fn(),
  updateStaffPropertySale: vi.fn(),
  updateStaffServicedApartment: vi.fn(),
}));

vi.mock('@/httpapis/uploads', () => ({
  completeUpload: vi.fn(),
  createUploadPresign: vi.fn(),
}));

interface EditorTestForm {
  addressText: string;
  addressTextEn: string;
  askingPriceHKD: number;
  contactAttributes: Record<string, string>;
  contactNameEn: string;
  contactNameZh: string;
  description: string;
  descriptionEn: string;
  districtCode: string;
  estateName: string;
  featureTags: string[];
  floorRaw: string;
  floorZone: string;
  locationAreaCode: string;
  locationDistrictCode: string;
  lowestMonthlyRentHKD: number;
  monthlyRentHKD: number;
  phone: string;
  phone2: string;
  priceNegotiable: boolean;
  priceReferenceOnly: boolean;
  propertyNo: string;
  propertyAttributes: Record<string, string>;
  propertyType: string;
  projectAttributes: Record<string, string>;
  projectName: string;
  publisherIdentityType: string;
  roomTypes: Array<{
    min_stay_value: number;
    name: string;
    usable_area_min_sqft: number;
  }>;
  serviceWhatsApp: string;
  summary: string;
  title: string;
  titleEn: string;
  transactionType: 'sale' | 'rent';
  usableAreaSqft: number;
  wechat: string;
}

interface EditorTestState {
  activeEditorStep: string;
  editorSteps: Array<{ key: string; label: string }>;
  form: EditorTestForm;
  images: Array<{
    id: string;
    isCover: boolean;
    mediaAssetId?: string;
    uploading: boolean;
    url?: string;
  }>;
  loadedPublicationStatus: string;
}

const mockedCreatePropertySale = vi.mocked(createPropertySale);
const mockedCreateServicedApartment = vi.mocked(createServicedApartment);
const mockedUpdatePropertySale = vi.mocked(updatePropertySale);
const mockedPublishPropertySale = vi.mocked(publishPropertySale);
const mockedPublishServicedApartment = vi.mocked(publishServicedApartment);
const mockedTranslatePropertyContent = vi.mocked(translatePropertyContent);

// 1. 建立樓盤編輯頁
const mountEditor = (
  channel: 'sale' | 'serviced' = 'sale',
  showProgress = false,
  listingId = '',
) => mount(PropertyEditorPage, {
  props: {
    channel,
    embedded: true,
    hideHeader: true,
    hideProgress: !showProgress,
    listingId,
  },
  global: {
    stubs: {
      AppIcon: true,
      AppUnsavedChangesDialog: true,
    },
  },
});

// 2. 填入可發布的住宅樓盤資料
const fillValidSale = (state: EditorTestState): void => {
  Object.assign(state.form, {
    addressText: '皇后大道中1號',
    addressTextEn: "1 Queen's Road Central",
    askingPriceHKD: 680,
    contactAttributes: {
      agency_company_profile: 'AJO Property Agency Limited',
      agency_contact_profile: 'Agent Chan E-123456',
    },
    contactNameEn: 'Mr Chan',
    contactNameZh: '陳先生',
    description: '實用住宅單位，交通方便。',
    descriptionEn: 'Practical residential unit with convenient transport.',
    districtCode: 'central',
    estateName: '中環大廈',
    featureTags: ['residential_private_estate'],
    floorRaw: '25/F',
    floorZone: 'high',
    locationAreaCode: 'hong_kong_island',
    locationDistrictCode: 'central_western',
    monthlyRentHKD: 0,
    phone: '61234567',
    phone2: '62345678',
    priceNegotiable: false,
    priceReferenceOnly: true,
    propertyNo: 'AJO-1001',
    propertyAttributes: {
      account_or_package: 'personal_account',
      new_completion: 'yes',
    },
    publisherIdentityType: 'agent',
    title: '中環住宅放售',
    titleEn: 'Central residential property for sale',
    transactionType: 'sale',
    usableAreaSqft: 420,
    wechat: 'ajo-owner',
  });
  state.images = [{
    id: 'media-1',
    mediaAssetId: 'media-1',
    url: 'https://cdn.test/property.jpg',
    isCover: true,
    uploading: false,
  }];
  state.activeEditorStep = 'contact';
};

// 3. 點擊發布按鈕
const clickPublish = async (wrapper: ReturnType<typeof mountEditor>): Promise<void> => {
  await nextTick();
  const publishButton = wrapper.findAll('button').find((button) => button.text() === '儲存並發布');
  expect(publishButton).toBeTruthy();
  await publishButton?.trigger('click');
  await flushPromises();
};

// 4. 點擊保存草稿按鈕
const clickSaveDraft = async (wrapper: ReturnType<typeof mountEditor>): Promise<void> => {
  await nextTick();
  const saveDraftButton = wrapper.findAll('button').find((button) => button.text() === '保存草稿');
  expect(saveDraftButton).toBeTruthy();
  await saveDraftButton?.trigger('click');
  await flushPromises();
};

// 5. 取得測試用 setup state
const getEditorState = (wrapper: ReturnType<typeof mountEditor>): EditorTestState =>
  (wrapper.vm.$ as unknown as { setupState: EditorTestState }).setupState;

describe('PropertyEditorPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedCreatePropertySale.mockResolvedValue({
      data: { data: { listing_id: 'property-1' } },
    } as Awaited<ReturnType<typeof createPropertySale>>);
    mockedUpdatePropertySale.mockResolvedValue({
      data: { data: { listing_id: 'property-1' } },
    } as Awaited<ReturnType<typeof updatePropertySale>>);
    mockedPublishPropertySale.mockResolvedValue({
      data: { data: { listing_id: 'property-1', publication_status: 'active' } },
    } as Awaited<ReturnType<typeof publishPropertySale>>);
    mockedTranslatePropertyContent.mockResolvedValue({
      data: {
        data: {
          title_en: 'High-floor two-bedroom unit in Kornhill',
          description_en: 'Practical two-bedroom unit with convenient transport.',
        },
      },
    } as Awaited<ReturnType<typeof translatePropertyContent>>);
    testMocks.loadCurrentUser.mockResolvedValue(undefined);
  });

  it('creates, updates, and publishes a complete property sale', async () => {
    const wrapper = mountEditor();
    await flushPromises();
    const state = getEditorState(wrapper);
    fillValidSale(state);

    const identityField = wrapper.findAll('.property-input')
      .find((field) => field.text().includes('業主或地產代理'));
    expect(identityField).toBeTruthy();
    expect(identityField?.find('select').exists()).toBe(false);
    expect(identityField?.find('small').exists()).toBe(false);
    expect(identityField?.text()).toContain('業主');
    expect(wrapper.text()).not.toContain('身份按帳戶資料顯示');

    await clickPublish(wrapper);

    expect(mockedCreatePropertySale).toHaveBeenCalledOnce();
    expect(mockedCreatePropertySale).toHaveBeenCalledWith(expect.objectContaining({
      floor_raw: '25/F',
      floor_zone: 'high',
      price_reference_only: true,
      property_attributes: expect.not.objectContaining({ new_completion: 'yes' }),
    }));
    expect(mockedUpdatePropertySale).toHaveBeenCalledOnce();
    expect(mockedUpdatePropertySale).toHaveBeenCalledWith(
      'property-1',
      expect.objectContaining({
        images: [{ media_asset_id: 'media-1', sort_order: 1, is_cover: true }],
      }),
    );
    expect(mockedPublishPropertySale).toHaveBeenCalledWith('property-1');
    expect(wrapper.emitted('published')).toEqual([['property-1']]);
  });

  it('saves changes to an active property without publishing the draft again', async () => {
    const wrapper = mountEditor('sale', false, 'property-1');
    await flushPromises();
    const state = getEditorState(wrapper);
    fillValidSale(state);
    state.loadedPublicationStatus = 'active';
    state.form.contactNameZh = '';
    state.form.contactNameEn = '';
    state.form.phone = '';
    await nextTick();
    testMocks.pushToast.mockClear();

    expect(wrapper.find('.property-editor-charge').exists()).toBe(false);
    expect(wrapper.text()).not.toContain('保存草稿');
    const saveChangesButton = wrapper.findAll('button')
      .find((button) => button.text() === '儲存修改');
    expect(saveChangesButton).toBeTruthy();
    await saveChangesButton?.trigger('click');
    await flushPromises();

    expect(mockedUpdatePropertySale).toHaveBeenCalledWith(
      'property-1',
      expect.objectContaining({
        images: [{ media_asset_id: 'media-1', sort_order: 1, is_cover: true }],
      }),
    );
    expect(mockedPublishPropertySale).not.toHaveBeenCalled();
    expect(wrapper.emitted('saved')).toEqual([['property-1']]);
    expect(testMocks.pushToast).toHaveBeenCalledWith('樓盤修改已儲存', 'success');
  });

  it('saves an incomplete property sale as a draft without publishing', async () => {
    const wrapper = mountEditor();
    await flushPromises();
    const state = getEditorState(wrapper);
    state.activeEditorStep = 'contact';

    await clickSaveDraft(wrapper);

    expect(mockedCreatePropertySale).toHaveBeenCalledOnce();
    expect(mockedCreatePropertySale).toHaveBeenCalledWith(
      expect.objectContaining({ images: [] }),
      { charge_draft: true },
    );
    expect(mockedUpdatePropertySale).toHaveBeenCalledWith(
      'property-1',
      expect.objectContaining({ images: [] }),
    );
    expect(mockedPublishPropertySale).not.toHaveBeenCalled();
    expect(wrapper.emitted('saved')).toEqual([['property-1']]);
  });

  it('saves separate WhatsApp settings for phone 1 and phone 2', async () => {
    const wrapper = mountEditor();
    await flushPromises();
    const state = getEditorState(wrapper);
    fillValidSale(state);
    state.form.publisherIdentityType = 'owner';
    state.form.contactAttributes.phone_country_code = '+852';
    state.form.contactAttributes.phone_2_country_code = '+86';
    state.form.contactAttributes.phone_whatsapp_enabled = 'yes';
    state.form.contactAttributes.phone_2_whatsapp_enabled = 'yes';
    state.activeEditorStep = 'contact';
    await nextTick();

    expect(wrapper.text().match(/可 WhatsApp 聯絡/g)).toHaveLength(2);
    expect(wrapper.text()).toContain('Wechat ID');
    expect(wrapper.findAll('.property-phone-group__code')).toHaveLength(2);
    expect((wrapper.findAll('.property-phone-group__code')[1]?.element as HTMLSelectElement).value).toBe('+86');

    await clickPublish(wrapper);

    expect(mockedCreatePropertySale).toHaveBeenCalledWith(expect.objectContaining({
      contact: expect.objectContaining({
        phone: '61234567',
        phone_2: '62345678',
        whatsapp: '61234567',
        wechat: 'ajo-owner',
        show_whatsapp: true,
        contact_attributes: expect.objectContaining({
          phone_country_code: '+852',
          phone_2_country_code: '+86',
          phone_whatsapp_enabled: 'yes',
          phone_2_whatsapp_enabled: 'yes',
        }),
      }),
    }));
  });

  it('shows sale or rent pricing copy within the four-step publishing flow', async () => {
    const wrapper = mountEditor();
    await flushPromises();
    const state = getEditorState(wrapper);
    state.activeEditorStep = 'details';
    await nextTick();

    expect(state.editorSteps).toHaveLength(4);
    expect(wrapper.text()).toContain('售價');
    expect(wrapper.text()).toContain('標示的價格為議價基礎');
    expect(wrapper.text()).not.toContain('價格及租期');

    state.form.transactionType = 'rent';
    state.form.monthlyRentHKD = 18000;
    await nextTick();

    expect(wrapper.text()).toContain('租金');
    expect(wrapper.text()).not.toContain('標示的價格為議價基礎');
  });

  it('does not submit hidden sale-only pricing flags after switching to rent', async () => {
    const wrapper = mountEditor();
    await flushPromises();
    const state = getEditorState(wrapper);
    fillValidSale(state);
    state.form.priceNegotiable = true;
    state.form.transactionType = 'rent';
    state.form.monthlyRentHKD = 18000;

    await clickPublish(wrapper);

    expect(mockedCreatePropertySale).toHaveBeenCalledWith(expect.objectContaining({
      monthly_rent_hkd: 18000,
      price_negotiable: false,
      price_reference_only: false,
    }));
  });

  it('translates the Chinese title and description into separate English rows', async () => {
    const wrapper = mountEditor();
    await flushPromises();
    const state = getEditorState(wrapper);
    state.activeEditorStep = 'details';
    state.form.title = '康怡花園高層兩房';
    state.form.description = '實用兩房，交通方便。';
    state.form.titleEn = '';
    state.form.descriptionEn = '';
    await nextTick();

    const translationButton = wrapper.findAll('button')
      .find((button) => button.text() === '翻譯成 English');
    expect(translationButton).toBeTruthy();
    await translationButton?.trigger('click');
    await flushPromises();

    expect(mockedTranslatePropertyContent).toHaveBeenCalledWith({
      title: '康怡花園高層兩房',
      description: '實用兩房，交通方便。',
    });
    expect(state.form.titleEn).toBe('High-floor two-bedroom unit in Kornhill');
    expect(state.form.descriptionEn).toBe('Practical two-bedroom unit with convenient transport.');
    expect(testMocks.pushToast).toHaveBeenCalledWith('已產生 English 內容', 'success');

    const bilingualFields = wrapper.findAll('.property-section-heading + .property-editor-grid > .property-input')
      .slice(0, 4);
    expect(bilingualFields).toHaveLength(4);
    bilingualFields.forEach((field) => expect(field.classes()).toContain('property-input--wide'));
  });

  it('keeps existing English content when translation fails', async () => {
    mockedTranslatePropertyContent.mockRejectedValueOnce(new Error('translation failed'));
    const wrapper = mountEditor();
    await flushPromises();
    const state = getEditorState(wrapper);
    state.activeEditorStep = 'details';
    state.form.title = '康怡花園高層兩房';
    state.form.description = '實用兩房，交通方便。';
    state.form.titleEn = 'Existing English title';
    state.form.descriptionEn = 'Existing English description.';
    await nextTick();

    const translationButton = wrapper.findAll('button')
      .find((button) => button.text() === '翻譯成 English');
    await translationButton?.trigger('click');
    await flushPromises();

    expect(state.form.titleEn).toBe('Existing English title');
    expect(state.form.descriptionEn).toBe('Existing English description.');
    expect(testMocks.pushToast).toHaveBeenCalledWith('暫時無法翻譯，請稍後再試。', 'error');
    expect(wrapper.text()).not.toContain('FAIL');
  });

  it('shows the missing field and does not create a draft', async () => {
    const wrapper = mountEditor();
    await flushPromises();
    const state = getEditorState(wrapper);
    fillValidSale(state);
    state.form.titleEn = '';

    await clickPublish(wrapper);

    expect(mockedCreatePropertySale).not.toHaveBeenCalled();
    expect(mockedPublishPropertySale).not.toHaveBeenCalled();
    expect(testMocks.pushToast).toHaveBeenCalledWith('請先填寫：放盤標題(英文)', 'error');
    expect(state.activeEditorStep).toBe('details');
    expect(wrapper.text()).toContain('請檢查以下必填欄位');
    const invalidField = wrapper.find('.property-validation-error');
    expect(invalidField.text()).toContain('放盤標題(英文)');
    expect(invalidField.attributes('data-property-validation-message')).toBe('請填寫「放盤標題(英文)」');
    expect(invalidField.find('input').attributes('aria-invalid')).toBe('true');
  });

  it('lists every missing field across steps and locates the selected input', async () => {
    const wrapper = mountEditor('sale', true);
    await flushPromises();
    const state = getEditorState(wrapper);
    fillValidSale(state);
    state.form.titleEn = '';
    state.form.contactNameEn = '';
    state.form.phone = '';

    await clickPublish(wrapper);

    expect(testMocks.pushToast).toHaveBeenCalledWith(
      '請先填寫：放盤標題(英文)、英文名、電話1',
      'error',
    );
    const summary = wrapper.find('.property-validation-summary');
    expect(summary.text()).toContain('放盤標題(英文)');
    expect(summary.text()).toContain('英文名');
    expect(summary.text()).toContain('電話1');
    expect(wrapper.findAll('.property-editor-progress__error-count').map((item) => item.text())).toEqual(['1', '2']);

    const phoneIssue = summary.findAll('button').find((button) => button.text().includes('電話1'));
    await phoneIssue?.trigger('click');
    await nextTick();

    expect(state.activeEditorStep).toBe('contact');
    const phoneField = wrapper.findAll('.property-validation-error')
      .find((field) => field.text().includes('電話1'));
    expect(phoneField?.attributes('data-property-validation-message')).toBe('請填寫「電話1」');
  });

  it('points to the exact incomplete serviced residence room field', async () => {
    const wrapper = mountEditor('serviced');
    await flushPromises();
    const state = getEditorState(wrapper);
    Object.assign(state.form, {
      districtCode: 'central',
      locationAreaCode: 'hong_kong_island',
      locationDistrictCode: 'central_western',
      lowestMonthlyRentHKD: 28000,
      phone: '61234567',
      projectAttributes: {
        address_street: '皇后大道中1號',
        address_doorplate: '',
      },
      projectName: '中環服務式住宅',
      serviceWhatsApp: '',
      summary: '適合短期入住的服務式住宅。',
    });
    state.form.roomTypes[0].name = '';
    state.form.roomTypes[0].usable_area_min_sqft = 320;
    state.form.roomTypes[0].min_stay_value = 1;
    state.images = [{
      id: 'media-1',
      mediaAssetId: 'media-1',
      url: 'https://cdn.test/serviced.jpg',
      isCover: true,
      uploading: false,
    }];
    state.activeEditorStep = 'details';

    await clickPublish(wrapper);

    expect(mockedCreateServicedApartment).not.toHaveBeenCalled();
    expect(mockedPublishServicedApartment).not.toHaveBeenCalled();
    expect(testMocks.pushToast).toHaveBeenCalledWith(
      expect.stringContaining('房型名稱 1'),
      'error',
    );
    expect(state.activeEditorStep).toBe('details');
    const invalidField = wrapper.find('.property-validation-error');
    expect(invalidField.text()).toContain('房型名稱');
    expect(invalidField.find('input').attributes('aria-invalid')).toBe('true');
  });

  it('points to the exact incomplete field for non-residential property listings', async () => {
    const wrapper = mountEditor();
    await flushPromises();
    const state = getEditorState(wrapper);
    fillValidSale(state);
    state.form.propertyType = 'office';
    state.form.titleEn = '';

    await clickPublish(wrapper);

    expect(mockedCreatePropertySale).not.toHaveBeenCalled();
    expect(testMocks.pushToast).toHaveBeenCalledWith(
      expect.stringContaining('放盤標題(英文)'),
      'error',
    );
    const invalidField = wrapper.find('.property-validation-error');
    expect(invalidField.text()).toContain('放盤標題(英文)');
    expect(invalidField.find('input').attributes('aria-invalid')).toBe('true');
  });

  it('requires a photo before creating and publishing a draft', async () => {
    const wrapper = mountEditor();
    await flushPromises();
    const state = getEditorState(wrapper);
    fillValidSale(state);
    state.images = [];

    await clickPublish(wrapper);

    expect(mockedCreatePropertySale).not.toHaveBeenCalled();
    expect(mockedPublishPropertySale).not.toHaveBeenCalled();
    expect(testMocks.pushToast).toHaveBeenCalledWith('請先填寫：相片', 'error');
    expect(state.activeEditorStep).toBe('details');
    expect(wrapper.find('.property-media-panel').classes()).toContain('property-validation-error');
  });

  it('shows the streamlined residential property fields in the requested order', async () => {
    const wrapper = mountEditor();
    await flushPromises();
    const state = getEditorState(wrapper);
    state.activeEditorStep = 'details';
    await nextTick();

    const propertyPanel = wrapper.findAll('.property-editor-panel')
      .find((panel) => panel.find('h2').text() === '物業資料');
    const detailsText = propertyPanel?.text() ?? '';
    const areaIndex = detailsText.indexOf('區域');
    const districtIndex = detailsText.indexOf('地區');
    const subdistrictIndex = detailsText.indexOf('分區');
    const buildingIndex = detailsText.indexOf('大廈名稱');
    const addressIndex = detailsText.indexOf('街名及門牌');
    const addressEnIndex = detailsText.indexOf('街名及門牌(英文)');
    const grossAreaIndex = detailsText.indexOf('建築面積');
    const usableAreaIndex = detailsText.indexOf('實用面積');
    const buildingAgeIndex = detailsText.indexOf('樓齡');
    const blockIndex = detailsText.indexOf('座數');
    const actualFloorIndex = detailsText.indexOf('實際樓層');
    const unitIndex = detailsText.indexOf('單位');
    const directionIndex = detailsText.indexOf('座向');
    const managementFeeIndex = detailsText.indexOf('管理費');
    const propertyFields = propertyPanel?.findAll('.property-editor-grid > .property-input') ?? [];

    expect(detailsText).toContain('物業資料');
    expect(districtIndex).toBeGreaterThan(areaIndex);
    expect(subdistrictIndex).toBeGreaterThan(districtIndex);
    expect(buildingIndex).toBeGreaterThan(subdistrictIndex);
    expect(addressIndex).toBeGreaterThan(buildingIndex);
    expect(addressEnIndex).toBeGreaterThan(addressIndex);
    expect(grossAreaIndex).toBeGreaterThan(addressEnIndex);
    expect(usableAreaIndex).toBeGreaterThan(grossAreaIndex);
    expect(buildingAgeIndex).toBeGreaterThan(usableAreaIndex);
    expect(blockIndex).toBeGreaterThan(buildingAgeIndex);
    expect(actualFloorIndex).toBeGreaterThan(blockIndex);
    expect(unitIndex).toBeGreaterThan(actualFloorIndex);
    expect(directionIndex).toBeGreaterThan(unitIndex);
    expect(managementFeeIndex).toBeGreaterThan(directionIndex);
    expect(detailsText.match(/座向/g)).toHaveLength(1);
    expect(detailsText.match(/管理費/g)).toHaveLength(1);
    expect(detailsText).not.toContain('新落成樓盤');
    expect(detailsText).not.toContain('顯示樓層');
    expect(wrapper.find('.property-floor-unit-fields').exists()).toBe(true);
    expect(wrapper.findAll('.property-floor-unit-fields > .property-input')).toHaveLength(2);
    expect(propertyPanel).toBeTruthy();
    expect(propertyFields[propertyFields.length - 1]?.text()).toContain('管理費');
  });

  it('keeps applicable property fields in the same order for every non-residential listing type', async () => {
    const wrapper = mountEditor();
    await flushPromises();
    const state = getEditorState(wrapper);
    state.activeEditorStep = 'details';

    const cases = [
      {
        type: 'car_park',
        labels: ['區域', '地區', '分區', '大廈名稱', '街名及門牌', '街名及門牌(英文)', '樓齡', '座數 / 大廈', '實際樓層', '單位', '管理費'],
      },
      {
        type: 'industrial',
        labels: ['區域', '地區', '分區', '工商大廈名稱', '街名及門牌', '街名及門牌(英文)', '建築面積', '實用面積', '樓齡', '座數 / 大廈', '實際樓層', '單位', '座向', '管理費'],
      },
      {
        type: 'shop',
        labels: ['區域', '地區', '分區', '工商大廈名稱', '街名及門牌', '街名及門牌(英文)', '建築面積', '實用面積', '樓齡', '座數 / 大廈', '實際樓層', '單位', '管理費'],
      },
      {
        type: 'land',
        labels: ['區域', '地區', '分區', '街名及門牌', '街名及門牌(英文)', '建築面積', '實用面積'],
      },
    ];

    for (const item of cases) {
      state.form.propertyType = item.type;
      await nextTick();

      const propertyPanel = wrapper.findAll('.property-editor-panel')
        .find((panel) => panel.find('h2').text() === '物業資料');
      const detailsText = propertyPanel?.text() ?? '';
      const titlePanel = wrapper.findAll('.property-editor-panel')
        .find((panel) => panel.find('h2').text() === 'property.editor.saleModuleD');

      item.labels.forEach((label, index) => {
        expect(detailsText.indexOf(label), `${item.type}: ${label}`).toBeGreaterThan(
          index === 0 ? -1 : detailsText.indexOf(item.labels[index - 1] ?? ''),
        );
      });
      expect(titlePanel?.text()).not.toContain('座向');
      expect(titlePanel?.text()).not.toContain('管理費');
    }
  });

  it('only shows residential details for selected categories and clears hidden values', async () => {
    const wrapper = mountEditor();
    await flushPromises();
    const state = getEditorState(wrapper);
    state.activeEditorStep = 'details';
    await nextTick();

    const selector = wrapper.find('.property-feature-group-selector');
    const groupInputs = selector.findAll('input[type="checkbox"]');

    expect(selector.text()).toContain('細則欄目');
    expect(groupInputs).toHaveLength(7);
    expect(wrapper.findAll('.property-feature-detail-group')).toHaveLength(0);

    await groupInputs[0]?.setValue(true);
    expect(wrapper.findAll('.property-feature-detail-group')).toHaveLength(1);
    expect(wrapper.find('.property-feature-detail-group').text()).toContain('望山景');

    const firstDetailInput = wrapper.find('.property-feature-detail-group input[type="checkbox"]');
    await firstDetailInput.setValue(true);
    expect(state.form.featureTags).toContain('view_mountain');

    await groupInputs[0]?.setValue(false);
    expect(wrapper.findAll('.property-feature-detail-group')).toHaveLength(0);
    expect(state.form.featureTags).not.toContain('view_mountain');

    state.form.featureTags = ['view_mountain'];
    await nextTick();
    expect(wrapper.findAll('.property-feature-detail-group')).toHaveLength(1);
    expect(groupInputs[0]?.element).toHaveProperty('checked', true);
  });
});
