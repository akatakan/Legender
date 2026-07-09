import { Champion, ChampionModal } from '../components/ChampionModal';
import { useLCU } from '../hooks/useLCU';
import { setAutoAcceptState, fetchSettings, updateAutoPickSettings } from '../services/lcu';
import { useEffect, useState } from 'react';

export const Home = () => {
    const { lcuData, isLoading, checkConnection } = useLCU();
    const [isAutoAcceptOn, setIsAutoAcceptOn] = useState(false);
    const [autoPick, setAutoPick] = useState({ enabled: false, primaryChamp: 0, secondaryChamp: 0, champPool: [] as number[] });
    const [modalConfig, setModalConfig] = useState<{ isOpen: boolean, type: 'primary' | 'secondary' | 'pool' }>({ isOpen: false, type: 'primary' });
    const [championsData, setChampionsData] = useState<Champion[]>([]);

    useEffect(() => {
        const init = async () => {
            const settings = await fetchSettings();
            setIsAutoAcceptOn(settings.autoAccept);
            setAutoPick(settings.autoPick);

            try {
                const vRes = await fetch('https://ddragon.leagueoflegends.com/api/versions.json');
                const versions = await vRes.json();
                const cRes = await fetch(`https://ddragon.leagueoflegends.com/cdn/${versions[0]}/data/tr_TR/champion.json`);
                const cData = await cRes.json();
                const champArray: Champion[] = Object.values(cData.data).map((c: any) => ({
                    id: parseInt(c.key),
                    name: c.name,
                    image: `https://ddragon.leagueoflegends.com/cdn/${versions[0]}/img/champion/${c.image.full}`
                }));
                setChampionsData(champArray);
            } catch (e) { console.error("Şampiyon listesi çekilemedi", e); }
        };
        init();
    }, []);

    const handleAutoPickUpdate = async (newConfig: any) => {
        setAutoPick(newConfig);
        await updateAutoPickSettings(newConfig);
    };

    const getChampUI = (id: number) => {
        const champ = championsData.find(c => c.id === id);
        if (!champ) return <span className="text-[#73829A] text-sm italic">Şampiyon Seç</span>;
        return (
            <div className="flex items-center gap-2">
                <img src={champ.image} alt={champ.name} className="w-8 h-8 rounded-full border border-[#D4AF37]" />
                <span className="text-[#E2E8F0] text-sm font-medium">{champ.name}</span>
            </div>
        );
    };

    const handleToggle = async () => {
        const newState = !isAutoAcceptOn;
        setIsAutoAcceptOn(newState);
        await setAutoAcceptState(newState);
    };

    return (
        <div className="min-h-screen bg-[#0B0E14] text-[#E2E8F0] p-8 font-sans selection:bg-[#D4AF37] selection:text-black">
            <div className="max-w-3xl mx-auto space-y-8">
                <header className="flex items-center justify-between pb-6 border-b border-[#1F242F]">
                    <div>
                        <h1 className="text-3xl font-light tracking-wide text-white">
                            LoL <span className="font-bold text-[#D4AF37]">Companion</span>
                        </h1>
                        <p className="text-sm text-[#73829A] mt-1">Premium LCU Assistant</p>
                    </div>
                    <button
                        onClick={checkConnection}
                        disabled={isLoading}
                        className={`px-6 py-2.5 rounded-md font-medium tracking-wide transition-all duration-300 ${isLoading ? 'bg-[#1F242F] text-[#73829A] cursor-not-allowed' : 'bg-[#D4AF37] hover:bg-[#F3D360] text-[#0B0E14] shadow-[0_0_15px_rgba(212,175,55,0.3)] hover:shadow-[0_0_25px_rgba(212,175,55,0.5)] active:scale-95'}`}
                    >
                        {isLoading ? 'Senkronize Ediliyor...' : 'Bağlantıyı Yenile'}
                    </button>
                </header>

                {lcuData && (
                    <div className="space-y-6">
                        <div className="bg-[#151923] border border-[#1F242F] rounded-xl overflow-hidden shadow-2xl relative">
                            <div className={`h-1 w-full ${lcuData.isActive ? 'bg-gradient-to-r from-[#10B981] to-[#047857]' : 'bg-gradient-to-r from-[#EF4444] to-[#B91C1C]'}`}></div>
                            <div className="p-8">
                                <div className="flex items-center justify-between mb-8">
                                    <div className="flex items-center gap-3">
                                        <div className="relative flex h-3 w-3">
                                            {lcuData.isActive && <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-[#10B981] opacity-75"></span>}
                                            <span className={`relative inline-flex rounded-full h-3 w-3 ${lcuData.isActive ? 'bg-[#10B981]' : 'bg-[#EF4444]'}`}></span>
                                        </div>
                                        <span className="text-sm font-medium tracking-wider uppercase text-[#94A3B8]">
                                            {lcuData.isActive ? 'İstemci Çevrimiçi' : 'İstemci Çevrimdışı'}
                                        </span>
                                    </div>
                                    {lcuData.isActive && <div className="text-xs font-mono text-[#475569] bg-[#0B0E14] px-3 py-1 rounded-md border border-[#1F242F]">PORT: {lcuData.port}</div>}
                                </div>

                                {lcuData.isActive && lcuData.summoner ? (
                                    <div className="flex items-center gap-8">
                                        <div className="relative group">
                                            <div className="absolute -inset-0.5 bg-gradient-to-b from-[#D4AF37] to-[#735F1E] rounded-full opacity-50 group-hover:opacity-100 transition duration-300 blur-sm"></div>
                                            <div className="relative w-28 h-28 rounded-full bg-[#0B0E14] p-1">
                                                <img src={`https://raw.communitydragon.org/latest/plugins/rcp-be-lol-game-data/global/default/v1/profile-icons/${lcuData.summoner.profileIconId}.jpg`} className="w-full h-full rounded-full object-cover border border-[#1F242F]" onError={(e) => { e.currentTarget.src = 'https://raw.communitydragon.org/latest/plugins/rcp-be-lol-game-data/global/default/v1/profile-icons/0.jpg' }} />
                                            </div>
                                            <div className="absolute -bottom-3 left-1/2 -translate-x-1/2 bg-[#0B0E14] border border-[#D4AF37] text-[#D4AF37] text-xs font-bold px-3 py-1 rounded-full shadow-lg">Lvl {lcuData.summoner.summonerLevel}</div>
                                        </div>
                                        <div className="flex flex-col justify-center">
                                            <h2 className="text-3xl font-bold text-white tracking-tight mb-2">{lcuData.summoner.displayName || "Sihirdar"}</h2>
                                            <p className="text-[#73829A] text-sm font-medium flex items-center gap-2"><span>League of Legends</span><span className="w-1 h-1 rounded-full bg-[#475569]"></span><span>Bağlantı Kuruldu</span></p>
                                        </div>
                                    </div>
                                ) : (
                                    lcuData.isActive && <div className="text-[#D4AF37] bg-[#D4AF37]/10 p-5 rounded-lg border border-[#D4AF37]/20">Oyuncu verileri yüklenemedi. Lobi ekranında olduğunuzdan emin olun.</div>
                                )}
                            </div>
                        </div>

                        {lcuData.isActive && (
                            <div className="grid grid-cols-2 gap-4 mt-6">
                                {/* Otomatik Kabul Modülü */}
                                <div className="bg-[#151923] border border-[#1F242F] rounded-xl p-6 flex flex-col justify-between">
                                    <h3 className="text-sm tracking-wider uppercase text-[#73829A] font-semibold mb-4">Otomasyon</h3>
                                    <div className="bg-[#0B0E14] border border-[#1F242F] p-4 rounded-lg flex items-center justify-between">
                                        <span className={isAutoAcceptOn ? 'text-[#D4AF37]' : 'text-[#73829A]'}>Otomatik Kabul</span>
                                        <button onClick={handleToggle} className={`h-6 w-11 rounded-full transition-colors ${isAutoAcceptOn ? 'bg-[#D4AF37]' : 'bg-[#1F242F]'}`}>
                                            <span className={`block h-4 w-4 bg-white rounded-full transition-transform ${isAutoAcceptOn ? 'translate-x-6' : 'translate-x-1'}`} />
                                        </button>
                                    </div>
                                </div>

                                {/* Hızlı Seçim Modülü */}
                                <div className="bg-[#151923] border border-[#1F242F] rounded-xl p-6">
                                    <div className="flex justify-between items-center mb-4">
                                        <h3 className="text-sm tracking-wider uppercase text-[#73829A] font-semibold">Hızlı Seçim</h3>
                                        <button onClick={() => handleAutoPickUpdate({ ...autoPick, enabled: !autoPick.enabled })} className={`h-5 w-9 rounded-full transition-colors ${autoPick.enabled ? 'bg-[#D4AF37]' : 'bg-[#1F242F]'}`}>
                                            <span className={`block h-3 w-3 bg-white rounded-full transition-transform ${autoPick.enabled ? 'translate-x-5' : 'translate-x-1'}`} />
                                        </button>
                                    </div>

                                    <div className="space-y-3">
                                        <button onClick={() => setModalConfig({ isOpen: true, type: 'primary' })} className="w-full flex items-center bg-[#0B0E14] border border-[#1F242F] rounded-md px-3 py-2 text-left">{getChampUI(autoPick.primaryChamp)}</button>
                                        <button onClick={() => setModalConfig({ isOpen: true, type: 'secondary' })} className="w-full flex items-center bg-[#0B0E14] border border-[#1F242F] rounded-md px-3 py-2 text-left">{getChampUI(autoPick.secondaryChamp)}</button>
                                        <button onClick={() => setModalConfig({ isOpen: true, type: 'pool' })} className="w-full flex items-center justify-between bg-[#0B0E14] border border-[#1F242F] rounded-md px-3 py-2 text-left">
                                            <span className={(autoPick?.champPool?.length || 0) > 0 ? "text-[#E2E8F0] text-sm" : "text-[#73829A] text-sm"}>
                                                {(autoPick?.champPool?.length || 0) > 0 ? `${autoPick.champPool.length} Şampiyon Seçildi` : 'Havuz Belirle'}
                                            </span>
                                        </button>
                                    </div>
                                </div>
                            </div>
                        )}
                    </div>
                )}
            </div>

            <ChampionModal
                isOpen={modalConfig.isOpen}
                isMulti={modalConfig.type === 'pool'}
                initialSelected={modalConfig.type === 'pool' ? autoPick.champPool : []}
                onClose={() => setModalConfig({ ...modalConfig, isOpen: false })}
                onSelect={(val) => {
                    const newConfig = { ...autoPick };
                    if (modalConfig.type === 'primary') newConfig.primaryChamp = val as number;
                    if (modalConfig.type === 'secondary') newConfig.secondaryChamp = val as number;
                    if (modalConfig.type === 'pool') newConfig.champPool = val as number[];
                    handleAutoPickUpdate(newConfig);
                }}
            />
        </div>
    );
};