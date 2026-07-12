/*
 * 物業發布編輯頁測試。
 * 1. 驗證完整樓盤資料會依序建立草稿、更新圖片並發布。
 * 2. 驗證未完成資料可以保存草稿但不會發布。
 * 3. 驗證缺少必填欄位時會指出欄位並阻止發布請求。
 * 4. 驗證缺少相片時會返回資料步驟並阻止發布。
 */
import { flushPromises, mount } from '@vue/test-utils';
import { nextTick } from 'vue';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import {
  createPropertySale,
  publishPropertySale,
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
        'property.editor.imageRequired': '發布前至少需要一張相片',
        'property.editor.publishNow': '儲存並發布',
        'property.editor.publishSuccess': '已發布',
        'property.editor.saveDraft': '保存草稿',
        'property.editor.requiredFieldsDetail': '請先填寫：{fields}',
        'property.editor.titleEnField': '放盤標題(英文)',
      };

      return (messages[key] ?? key).replace('{fields}', String(params?.fields ?? ''));
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
      publisher_identity_type: 'owner',
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
  searchPropertyAddresses: vi.fn(),
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
  contactNameEn: string;
  contactNameZh: string;
  description: string;
  descriptionEn: string;
  districtCode: string;
  estateName: string;
  featureTags: string[];
  phone: string;
  title: string;
  titleEn: string;
  usableAreaSqft: number;
}

interface EditorTestState {
  activeEditorStep: string;
  form: EditorTestForm;
  images: Array<{
    id: string;
    isCover: boolean;
    mediaAssetId?: string;
    uploading: boolean;
    url?: string;
  }>;
}

const mockedCreatePropertySale = vi.mocked(createPropertySale);
const mockedUpdatePropertySale = vi.mocked(updatePropertySale);
const mockedPublishPropertySale = vi.mocked(publishPropertySale);

// 1. 建立樓盤編輯頁
const mountEditor = () => mount(PropertyEditorPage, {
  props: {
    channel: 'sale',
    embedded: true,
    hideHeader: true,
    hideProgress: true,
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
    contactNameEn: 'Mr Chan',
    contactNameZh: '陳先生',
    description: '實用住宅單位，交通方便。',
    descriptionEn: 'Practical residential unit with convenient transport.',
    districtCode: 'central',
    estateName: '中環大廈',
    featureTags: ['residential_private_estate'],
    phone: '61234567',
    title: '中環住宅放售',
    titleEn: 'Central residential property for sale',
    usableAreaSqft: 420,
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
    testMocks.loadCurrentUser.mockResolvedValue(undefined);
  });

  it('creates, updates, and publishes a complete property sale', async () => {
    const wrapper = mountEditor();
    await flushPromises();
    const state = getEditorState(wrapper);
    fillValidSale(state);

    await clickPublish(wrapper);

    expect(mockedCreatePropertySale).toHaveBeenCalledOnce();
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

  it('saves an incomplete property sale as a draft without publishing', async () => {
    const wrapper = mountEditor();
    await flushPromises();
    const state = getEditorState(wrapper);
    state.activeEditorStep = 'contact';

    await clickSaveDraft(wrapper);

    expect(mockedCreatePropertySale).toHaveBeenCalledOnce();
    expect(mockedUpdatePropertySale).toHaveBeenCalledWith(
      'property-1',
      expect.objectContaining({ images: [] }),
    );
    expect(mockedPublishPropertySale).not.toHaveBeenCalled();
    expect(wrapper.emitted('saved')).toEqual([['property-1']]);
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
    expect(testMocks.pushToast).toHaveBeenCalledWith('發布前至少需要一張相片', 'error');
    expect(state.activeEditorStep).toBe('details');
  });
});
