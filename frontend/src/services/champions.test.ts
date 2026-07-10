import { describe, expect, it } from 'vitest';
import {
    communityDragonProfileIconSources,
    normalizeDataDragonLocale,
    normalizePatchVersion,
    selectDataDragonVersion,
} from './champions';

describe('champion asset context', () => {
    it('normalizes Data Dragon locales', () => {
        expect(normalizeDataDragonLocale('tr-tr')).toBe('tr_TR');
        expect(normalizeDataDragonLocale('EN_us')).toBe('en_US');
        expect(normalizeDataDragonLocale('invalid')).toBe('en_US');
    });

    it('matches the installed client patch to a Data Dragon version', () => {
        const versions = ['16.14.1', '16.13.1', '16.12.1'];
        expect(normalizePatchVersion('16.13.123.456')).toBe('16.13');
        expect(selectDataDragonVersion(versions, '16.13.123.456')).toBe('16.13.1');
        expect(selectDataDragonVersion(versions, '15.1.0')).toBe('16.14.1');
    });

    it('orders CommunityDragon locale and version fallbacks', () => {
        const sources = communityDragonProfileIconSources(7, '16.13.123.456', 'tr_TR');
        expect(sources[0]).toContain('/16.13/plugins/rcp-be-lol-game-data/global/tr_tr/v1/profile-icons/7.jpg');
        expect(sources[1]).toContain('/16.13/plugins/rcp-be-lol-game-data/global/default/v1/profile-icons/7.jpg');
        expect(sources[2]).toContain('/latest/plugins/rcp-be-lol-game-data/global/default/v1/profile-icons/7.jpg');
        expect(sources.at(-1)).toContain('/latest/plugins/rcp-be-lol-game-data/global/default/v1/profile-icons/0.jpg');
    });
});
