import { describe, expect, it } from 'vitest';

import {
  avatarKeyFromSeed,
  buildingAvatarKey,
  normalizeAvatarKey,
  presetAvatarUrl,
  resolveAvatarSrc,
} from './presetAvatars';

describe('presetAvatars', () => {
  it('normalizes catalog keys and rejects unknown values', () => {
    expect(normalizeAvatarKey(' AVO-03 ')).toBe('avo-03');
    expect(normalizeAvatarKey('avo-00')).toBe('');
    expect(normalizeAvatarKey('custom')).toBe('');
  });

  it('maps the same seed to the same catalog key', () => {
    expect(avatarKeyFromSeed('01KUSER001')).toBe(avatarKeyFromSeed('01KUSER001'));
    expect(normalizeAvatarKey(avatarKeyFromSeed('01KUSER001'))).not.toBe('');
    expect(avatarKeyFromSeed('01KUSER002')).not.toBe(avatarKeyFromSeed('01KUSER001'));
  });

  it('prefers a selected key, then an uploaded image, then a seed fallback', () => {
    expect(resolveAvatarSrc({
      avatarKey: 'avo-08',
      avatarUrl: 'https://cdn.example/custom.png',
      seed: 'user-1',
    })).toBe(presetAvatarUrl('avo-08'));
    expect(resolveAvatarSrc({
      avatarUrl: 'https://cdn.example/custom.png',
      seed: 'user-1',
    })).toBe('https://cdn.example/custom.png');
    expect(resolveAvatarSrc({ seed: 'user-1' })).toBe(presetAvatarUrl(avatarKeyFromSeed('user-1')));
    expect(buildingAvatarKey('0999900')).toBe(avatarKeyFromSeed('building:0999900'));
  });
});
