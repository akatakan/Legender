export interface Champion {
    id: number;
    name: string;
    image: string;
}

interface DataDragonChampion {
    key: string;
    name: string;
    image: { full: string };
}

interface CachedCatalog {
    savedAt: number;
    champions: Champion[];
}

const championRequests = new Map<string, Promise<Champion[]>>();
const catalogCachePrefix = 'legender:champions:';

export const fetchChampions = (gameVersion?: string, requestedLocale?: string): Promise<Champion[]> => {
    const locale = normalizeDataDragonLocale(requestedLocale);
    const cacheKey = `${normalizePatchVersion(gameVersion) ?? 'latest'}:${locale}`;
    const cached = championRequests.get(cacheKey);
    if (cached) return cached;

    const request = loadChampions(gameVersion, locale)
        .then((champions) => {
            writeCachedChampions(cacheKey, champions);
            return champions;
        })
        .catch((error) => {
            const cached = readCachedChampions(cacheKey) ?? findLocaleFallback(locale);
            if (cached) return cached;
            championRequests.delete(cacheKey);
            throw error;
        });
    championRequests.set(cacheKey, request);
    return request;
};

export const communityDragonProfileIconSources = (iconId: number, gameVersion?: string, requestedLocale?: string): string[] => {
    const version = normalizePatchVersion(gameVersion) ?? 'latest';
    const locale = normalizeCommunityDragonLocale(requestedLocale);
    const root = 'https://raw.communitydragon.org';
    const path = `plugins/rcp-be-lol-game-data/global`;
    const candidates = [
        `${root}/${version}/${path}/${locale}/v1/profile-icons/${iconId}.jpg`,
        `${root}/${version}/${path}/default/v1/profile-icons/${iconId}.jpg`,
        `${root}/latest/${path}/default/v1/profile-icons/${iconId}.jpg`,
        `${root}/latest/${path}/default/v1/profile-icons/0.jpg`,
    ];
    return [...new Set(candidates)];
};

const loadChampions = async (gameVersion: string | undefined, locale: string): Promise<Champion[]> => {
    const versionResponse = await fetch('https://ddragon.leagueoflegends.com/api/versions.json');
    if (!versionResponse.ok) {
        throw new Error(`Data Dragon sürüm isteği başarısız: ${versionResponse.status}`);
    }
    const versions: string[] = await versionResponse.json();
    const version = selectDataDragonVersion(versions, gameVersion);
    if (!version) {
        throw new Error('Data Dragon kullanılabilir sürüm döndürmedi');
    }

    const payload = await fetchChampionPayload(version, locale);
    return Object.values(payload.data)
        .map((champion) => ({
            id: Number.parseInt(champion.key, 10),
            name: champion.name,
            image: `https://ddragon.leagueoflegends.com/cdn/${version}/img/champion/${champion.image.full}`,
        }))
        .sort((left, right) => left.name.localeCompare(right.name, locale.replace('_', '-')));
};

const fetchChampionPayload = async (version: string, locale: string): Promise<{ data: Record<string, DataDragonChampion> }> => {
    const locales = locale === 'en_US' ? [locale] : [locale, 'en_US'];
    for (const candidate of locales) {
        const response = await fetch(`https://ddragon.leagueoflegends.com/cdn/${version}/data/${candidate}/champion.json`);
        if (response.ok) return response.json();
    }
    throw new Error(`Data Dragon şampiyon verisi bulunamadı (${version}, ${locale})`);
};

export const normalizeDataDragonLocale = (locale?: string): string => {
    const parts = (locale || 'tr_TR').replace('-', '_').split('_');
    if (parts.length !== 2) return 'en_US';
    return `${parts[0].toLowerCase()}_${parts[1].toUpperCase()}`;
};

const normalizeCommunityDragonLocale = (locale?: string): string => normalizeDataDragonLocale(locale).toLowerCase();

export const normalizePatchVersion = (version?: string): string | undefined => {
    const match = version?.match(/^(\d+)\.(\d+)/);
    return match ? `${match[1]}.${match[2]}` : undefined;
};

export const selectDataDragonVersion = (versions: string[], gameVersion?: string): string | undefined => {
    const patchVersion = normalizePatchVersion(gameVersion);
    return versions.find((candidate) => candidate === gameVersion || (patchVersion !== undefined && candidate.startsWith(`${patchVersion}.`))) ?? versions[0];
};

const writeCachedChampions = (cacheKey: string, champions: Champion[]): void => {
    try {
        window.localStorage.setItem(`${catalogCachePrefix}${cacheKey}`, JSON.stringify({ savedAt: Date.now(), champions }));
        pruneChampionCaches(5);
    } catch {
        // Cache yazılamaması canlı verinin kullanılmasını engellememeli.
    }
};

const readCachedChampions = (cacheKey: string): Champion[] | null => {
    return readCachedCatalog(cacheKey)?.champions ?? null;
};

const readCachedCatalog = (cacheKey: string): CachedCatalog | null => {
    try {
        const raw = window.localStorage.getItem(`${catalogCachePrefix}${cacheKey}`);
        if (!raw) return null;
        const parsed: Partial<CachedCatalog> = JSON.parse(raw);
        if (typeof parsed.savedAt !== 'number' || !Array.isArray(parsed.champions) || !parsed.champions.every(isChampion)) return null;
        return { savedAt: parsed.savedAt, champions: parsed.champions };
    } catch {
        return null;
    }
};

const findLocaleFallback = (locale: string): Champion[] | null => {
    try {
        const suffix = `:${locale}`;
		let newest: CachedCatalog | null = null;
        for (let index = window.localStorage.length - 1; index >= 0; index--) {
            const key = window.localStorage.key(index);
            if (!key?.startsWith(catalogCachePrefix) || !key.endsWith(suffix)) continue;
            const cacheKey = key.slice(catalogCachePrefix.length);
			const cached = readCachedCatalog(cacheKey);
			if (cached && (!newest || cached.savedAt > newest.savedAt)) newest = cached;
        }
		return newest?.champions ?? null;
    } catch {
        return null;
    }
    return null;
};

const pruneChampionCaches = (keep: number): void => {
    const catalogs: Array<{ key: string; savedAt: number }> = [];
    for (let index = 0; index < window.localStorage.length; index++) {
        const key = window.localStorage.key(index);
        if (!key?.startsWith(catalogCachePrefix)) continue;
        const cached = readCachedCatalog(key.slice(catalogCachePrefix.length));
        catalogs.push({ key, savedAt: cached?.savedAt ?? 0 });
    }
    catalogs
        .sort((left, right) => right.savedAt - left.savedAt)
        .slice(keep)
        .forEach(({ key }) => window.localStorage.removeItem(key));
};

const isChampion = (value: Champion): boolean => (
    typeof value?.id === 'number'
    && typeof value?.name === 'string'
    && typeof value?.image === 'string'
);
