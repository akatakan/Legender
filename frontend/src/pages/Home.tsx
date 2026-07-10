import { useEffect, useState } from 'react';
import { config } from '../../wailsjs/go/models';
import { ChampionModal } from '../components/ChampionModal';
import { FallbackImage } from '../components/FallbackImage';
import { TitleBar } from '../components/TitleBar';
import { useLCU } from '../hooks/useLCU';
import { Champion, communityDragonProfileIconSources, fetchChampions } from '../services/champions';
import { exportDiagnostics, fetchSettings, setAutoAcceptState, updateAutoPickSettings } from '../services/lcu';

type ModalType = 'primary' | 'secondary' | 'pool';

export const Home = () => {
    const { lcuData, isLoading, error: lcuError, checkConnection, clearError: clearLCUError } = useLCU();
    const [isAutoAcceptOn, setIsAutoAcceptOn] = useState(false);
    const [autoPick, setAutoPick] = useState({ enabled: false, primaryChamp: 0, secondaryChamp: 0, champPool: [] as number[] });
    const [modalConfig, setModalConfig] = useState<{ isOpen: boolean; type: ModalType }>({ isOpen: false, type: 'primary' });
    const [champions, setChampions] = useState<Champion[]>([]);
    const [errors, setErrors] = useState<{ settings?: string; champions?: string; action?: string }>({});

    useEffect(() => {
        fetchSettings()
            .then((settings) => {
                setIsAutoAcceptOn(settings.autoAccept);
                setAutoPick(settings.autoPick);
                setErrors((current) => ({ ...current, settings: undefined }));
            })
            .catch((error) => {
                console.error('Ayarlar yüklenemedi', error);
                setErrors((current) => ({ ...current, settings: 'Ayarlar yüklenemedi; varsayılan değerler gösteriliyor.' }));
            });
    }, []);

    useEffect(() => {
        let cancelled = false;
        const locale = lcuData?.locale || navigator.language || 'tr_TR';
        fetchChampions(lcuData?.gameVersion, locale)
            .then((result) => {
                if (!cancelled) {
                    setChampions(result);
                    setErrors((current) => ({ ...current, champions: undefined }));
                }
            })
            .catch((error) => {
                console.error('Şampiyon listesi çekilemedi', error);
                if (!cancelled) setErrors((current) => ({ ...current, champions: 'Şampiyon verileri indirilemedi. İnternet bağlantısını kontrol edin.' }));
            });
        return () => { cancelled = true; };
    }, [lcuData?.gameVersion, lcuData?.locale]);

    const updateAutoPick = async (next: config.AutoPickSettings) => {
        const previous = autoPick;
        setAutoPick(next);
        try {
            await updateAutoPickSettings(next);
            setErrors((current) => ({ ...current, action: undefined }));
        } catch (error) {
            setAutoPick(previous);
            console.error('Hızlı seçim ayarları kaydedilemedi', error);
            setErrors((current) => ({ ...current, action: 'Hızlı seçim ayarları kaydedilemedi.' }));
        }
    };

    const toggleAutoAccept = async () => {
        const next = !isAutoAcceptOn;
        setIsAutoAcceptOn(next);
        try {
            await setAutoAcceptState(next);
            setErrors((current) => ({ ...current, action: undefined }));
        } catch (error) {
            setIsAutoAcceptOn(!next);
            console.error('Otomatik kabul ayarı kaydedilemedi', error);
            setErrors((current) => ({ ...current, action: 'Otomatik kabul ayarı kaydedilemedi.' }));
        }
    };

    const exportLogs = async () => {
        try {
            await exportDiagnostics();
        } catch (error) {
            console.error('Tanılama logu dışa aktarılamadı', error);
            setErrors((current) => ({ ...current, action: 'Tanılama logu dışa aktarılamadı.' }));
        }
    };

    const selectedChampion = (id: number, emptyLabel: string) => {
        const champion = champions.find((item) => item.id === id);
        if (!champion) return <span className="text-sm text-[#7F8A9A]">{emptyLabel}</span>;
        return (
            <span className="flex min-w-0 items-center gap-3">
                <img src={champion.image} alt="" className="h-8 w-8 shrink-0 rounded-md object-cover" />
                <span className="truncate text-sm font-medium text-[#E8EBF0]">{champion.name}</span>
            </span>
        );
    };

    const connected = Boolean(lcuData?.isActive);
    const visibleError = lcuError || errors.action || errors.settings || errors.champions;

    return (
        <div className="flex h-screen flex-col overflow-hidden bg-[#0D1117] text-[#E8EBF0]">
            <TitleBar connected={connected} />

            <main className="min-h-0 flex-1 overflow-y-auto">
                <div className="mx-auto w-full max-w-5xl px-6 py-7 lg:px-8">
                    <header className="mb-6 flex flex-wrap items-end justify-between gap-4">
                        <div>
                            <p className="mb-1 text-xs font-semibold uppercase tracking-[0.18em] text-[#C8A24A]">League companion</p>
                            <h1 className="text-2xl font-semibold tracking-tight text-white">Kontrol merkezi</h1>
                            <p className="mt-1 text-sm text-[#8994A4]">İstemci bağlantını ve oyun otomasyonlarını tek yerden yönet.</p>
                        </div>
                        <button type="button" onClick={checkConnection} disabled={isLoading} className="flat-button flat-button-secondary">
                            <svg className={isLoading ? 'animate-spin' : ''} viewBox="0 0 20 20" aria-hidden="true">
                                <path d="M15.4 6.2A6 6 0 1 0 16 11" /><path d="M15.4 2.8v3.8h-3.8" />
                            </svg>
                            {isLoading ? 'Kontrol ediliyor' : 'Bağlantıyı yenile'}
                        </button>
                    </header>

                    {visibleError && (
                        <div role="alert" className="mb-5 flex items-start justify-between gap-4 rounded-lg border border-[#713B42] bg-[#25171B] px-4 py-3 text-sm text-[#F2A7AF]">
                            <span>{visibleError}</span>
                            <div className="flex shrink-0 items-center gap-3">
                                <button type="button" onClick={exportLogs} className="font-medium text-[#F2BBC1] hover:text-white">Logu dışa aktar</button>
                                <button type="button" onClick={() => { setErrors({}); clearLCUError(); }} aria-label="Hata mesajını kapat" className="text-lg leading-none hover:text-white">×</button>
                            </div>
                        </div>
                    )}

                    <section aria-labelledby="client-heading" className="flat-panel mb-5">
                        <div className="flex flex-wrap items-center justify-between gap-3 border-b border-[#28313D] px-5 py-4">
                            <div className="flex items-center gap-3">
                                <span className={`grid h-9 w-9 place-items-center rounded-lg ${connected ? 'bg-[#173528] text-[#52D49A]' : 'bg-[#202630] text-[#778293]'}`}>
                                    <svg className="h-5 w-5 fill-none stroke-current stroke-[1.5]" viewBox="0 0 20 20" aria-hidden="true"><rect x="6.5" y="5.5" width="7" height="7" rx="1" /><path d="M4 8H2m16 0h-2M4 12H2m16 0h-2M8 4V2m4 2V2m-4 14v2m4-2v2" /></svg>
                                </span>
                                <div>
                                    <h2 id="client-heading" className="text-sm font-semibold text-[#E8EBF0]">League Client</h2>
                                    <p className="text-xs text-[#7F8A9A]">{connected ? 'Bağlantı aktif ve otomasyonlar hazır' : 'İstemci açıldığında otomatik bağlanacak'}</p>
                                </div>
                            </div>
                            <span className={`status-badge ${connected ? 'status-badge-online' : 'status-badge-offline'}`}>
                                <span className="h-1.5 w-1.5 rounded-full bg-current" />
                                {connected ? 'Çevrimiçi' : 'Çevrimdışı'}
                            </span>
                        </div>

                        {connected && lcuData?.summoner ? (
                            <div className="flex flex-wrap items-center gap-5 px-5 py-5">
                                <FallbackImage
                                    sources={communityDragonProfileIconSources(lcuData.summoner.profileIconId, lcuData.gameVersion, lcuData.locale)}
                                    alt={`${lcuData.summoner.displayName || 'Sihirdar'} profil ikonu`}
                                    className="h-16 w-16 rounded-xl border border-[#303A47] object-cover"
                                />
                                <div className="min-w-0 flex-1">
                                    <h3 className="truncate text-xl font-semibold text-white">{lcuData.summoner.displayName || 'Sihirdar'}</h3>
                                    <p className="mt-1 text-sm text-[#8994A4]">Seviye {lcuData.summoner.summonerLevel}</p>
                                </div>
                                <div className="flex flex-wrap gap-2 text-xs text-[#9AA4B2]">
                                    {lcuData.locale && <span className="metadata-chip">{lcuData.locale}</span>}
                                    {lcuData.gameVersion && <span className="metadata-chip">v{lcuData.gameVersion.split('+')[0]}</span>}
                                </div>
                            </div>
                        ) : (
                            <div className="px-5 py-7 text-sm text-[#7F8A9A]">
                                {connected ? 'Oyuncu bilgileri bekleniyor…' : 'League Client şu anda kapalı. Bu ekran bağlantıyı otomatik olarak takip eder.'}
                            </div>
                        )}
                    </section>

                    <section aria-labelledby="automation-heading">
                        <div className="mb-3 flex items-center justify-between">
                            <div>
                                <h2 id="automation-heading" className="text-sm font-semibold text-[#E8EBF0]">Otomasyonlar</h2>
                                <p className="mt-0.5 text-xs text-[#7F8A9A]">Ayarlar anında kaydedilir ve istemci bağlantısında uygulanır.</p>
                            </div>
                        </div>

                        <div className="grid gap-4 md:grid-cols-2">
                            <article className="flat-panel p-5">
                                <div className="mb-5 flex items-start justify-between gap-4">
                                    <div>
                                        <h3 className="text-sm font-semibold text-[#E8EBF0]">Otomatik kabul</h3>
                                        <p className="mt-1 max-w-xs text-xs leading-5 text-[#7F8A9A]">Hazır kontrolü geldiğinde senin adına kabul eder.</p>
                                    </div>
                                    <Toggle enabled={isAutoAcceptOn} onToggle={toggleAutoAccept} label="Otomatik kabul" />
                                </div>
                                <div className="flex items-center gap-2 border-t border-[#28313D] pt-4 text-xs text-[#7F8A9A]">
                                    <span className={`h-1.5 w-1.5 rounded-full ${isAutoAcceptOn ? 'bg-[#43C98B]' : 'bg-[#657080]'}`} />
                                    {isAutoAcceptOn ? 'Etkin' : 'Devre dışı'}
                                </div>
                            </article>

                            <article className="flat-panel p-5">
                                <div className="mb-4 flex items-start justify-between gap-4">
                                    <div>
                                        <h3 className="text-sm font-semibold text-[#E8EBF0]">Hızlı seçim</h3>
                                        <p className="mt-1 max-w-xs text-xs leading-5 text-[#7F8A9A]">Öncelik sırasına göre uygun şampiyonu kilitler.</p>
                                    </div>
                                    <Toggle enabled={autoPick.enabled} onToggle={() => updateAutoPick({ ...autoPick, enabled: !autoPick.enabled })} label="Hızlı seçim" />
                                </div>

                                <div className="space-y-2">
                                    <SelectionRow label="Birincil" onClick={() => setModalConfig({ isOpen: true, type: 'primary' })}>
                                        {selectedChampion(autoPick.primaryChamp, 'Şampiyon seç')}
                                    </SelectionRow>
                                    <SelectionRow label="İkincil" onClick={() => setModalConfig({ isOpen: true, type: 'secondary' })}>
                                        {selectedChampion(autoPick.secondaryChamp, 'Şampiyon seç')}
                                    </SelectionRow>
                                    <SelectionRow label="Havuz" onClick={() => setModalConfig({ isOpen: true, type: 'pool' })}>
                                        <span className="text-sm text-[#A8B1BE]">{autoPick.champPool.length > 0 ? `${autoPick.champPool.length} şampiyon` : 'Havuz oluştur'}</span>
                                    </SelectionRow>
                                </div>
                            </article>
                        </div>
                    </section>
                </div>
            </main>

            <ChampionModal
                isOpen={modalConfig.isOpen}
                champions={champions}
                isMulti={modalConfig.type === 'pool'}
                initialSelected={modalConfig.type === 'pool' ? autoPick.champPool : []}
                onClose={() => setModalConfig((current) => ({ ...current, isOpen: false }))}
                onSelect={(value) => {
                    const next = { ...autoPick };
                    if (modalConfig.type === 'primary') next.primaryChamp = value as number;
                    if (modalConfig.type === 'secondary') next.secondaryChamp = value as number;
                    if (modalConfig.type === 'pool') next.champPool = value as number[];
                    void updateAutoPick(next);
                }}
            />
        </div>
    );
};

const Toggle = ({ enabled, onToggle, label }: { enabled: boolean; onToggle: () => void; label: string }) => (
    <button type="button" role="switch" aria-checked={enabled} aria-label={label} onClick={onToggle} className={`flat-toggle ${enabled ? 'flat-toggle-enabled' : ''}`}>
        <span />
    </button>
);

const SelectionRow = ({ label, onClick, children }: { label: string; onClick: () => void; children: React.ReactNode }) => (
    <button type="button" onClick={onClick} className="selection-row">
        <span className="w-16 shrink-0 text-left text-xs font-medium text-[#6F7A8A]">{label}</span>
        <span className="min-w-0 flex-1 text-left">{children}</span>
        <svg className="h-4 w-4 shrink-0 text-[#596575]" viewBox="0 0 16 16" aria-hidden="true"><path d="m6 3.5 4.5 4.5L6 12.5" /></svg>
    </button>
);
