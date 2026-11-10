/*
 * 靜態預設頭像圖庫。
 * 1. 以短 key 對應 public/avatars 下的 SVG，不經後端或 OSS。
 * 2. 無選定值時用識別碼穩定映射到同一張圖。
 */

export const PRESET_AVATAR_COUNT = 16;

export const PRESET_AVATAR_KEYS = Array.from({ length: PRESET_AVATAR_COUNT }, (_, index) => {
  const value = index + 1;
  return `avo-${String(value).padStart(2, '0')}`;
});

export interface PresetAvatarOption {
  key: string;
  url: string;
}

// 1. 驗證頭像 key
export const normalizeAvatarKey = (value?: string | null): string => {
  const key = String(value ?? '').trim().toLowerCase();
  if (!/^avo-(0[1-9]|1[0-6])$/.test(key)) {
    return '';
  }
  return key;
};

// 2. 由識別碼穩定映射頭像
export const avatarKeyFromSeed = (seed?: string | null): string => {
  const value = String(seed ?? '').trim();
  if (!value) {
    return 'avo-01';
  }

  let hash = 0;
  for (let index = 0; index < value.length; index += 1) {
    hash = (hash * 33 + value.charCodeAt(index)) >>> 0;
  }
  const selected = (hash % PRESET_AVATAR_COUNT) + 1;
  return `avo-${String(selected).padStart(2, '0')}`;
};

// 3. 取得靜態頭像路徑
export const presetAvatarUrl = (key: string): string => `/avatars/${key}.svg`;

// 4. 優先使用已選 key，其次既有圖片，最後用識別碼映射
export const resolveAvatarSrc = (input: {
  avatarKey?: string | null;
  avatarUrl?: string | null;
  seed?: string | null;
}): string => {
  const selectedKey = normalizeAvatarKey(input.avatarKey);
  if (selectedKey) {
    return presetAvatarUrl(selectedKey);
  }

  const uploadedUrl = String(input.avatarUrl ?? '').trim();
  if (uploadedUrl) {
    return uploadedUrl;
  }

  return presetAvatarUrl(avatarKeyFromSeed(input.seed));
};

// 5. 大廈群聊頭像
export const buildingAvatarKey = (buildingID?: string | null): string => {
  const value = String(buildingID ?? '').trim();
  if (!value) {
    return avatarKeyFromSeed('building');
  }
  return avatarKeyFromSeed(`building:${value}`);
};

// 6. 頭像選擇清單
export const presetAvatarOptions: PresetAvatarOption[] = PRESET_AVATAR_KEYS.map((key) => ({
  key,
  url: presetAvatarUrl(key),
}));
