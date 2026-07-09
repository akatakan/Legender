import { useState, useEffect, useMemo } from 'react';

// Şampiyon veri modeli
export interface Champion {
    id: number;
    name: string;
    image: string;
}

interface ChampionModalProps {
    isOpen: boolean;
    onClose: () => void;
    onSelect: (id: number | number[]) => void;
    isMulti?: boolean;
    initialSelected?: number[];
}

export const ChampionModal = ({ isOpen, onClose, onSelect, isMulti = false, initialSelected = [] }: ChampionModalProps) => {
    const [champions, setChampions] = useState<Champion[]>([]);
    const [search, setSearch] = useState('');
    const [selectedMulti, setSelectedMulti] = useState<number[]>(initialSelected);

    // İlk açılışta Riot Data Dragon'dan güncel şampiyon listesini çek
    useEffect(() => {
        const fetchChampions = async () => {
            try {
                // En güncel yama sürümünü al
                const vRes = await fetch('https://ddragon.leagueoflegends.com/api/versions.json');
                const versions = await vRes.json();
                const latest = versions[0];

                // Şampiyon verilerini çek (Türkçe)
                const cRes = await fetch(`https://ddragon.leagueoflegends.com/cdn/${latest}/data/tr_TR/champion.json`);
                const cData = await cRes.json();

                // Objeyi diziye çevir
                const champArray: Champion[] = Object.values(cData.data).map((c: any) => ({
                    id: parseInt(c.key),
                    name: c.name,
                    image: `https://ddragon.leagueoflegends.com/cdn/${latest}/img/champion/${c.image.full}`
                }));

                // Harf sırasına göre diz
                setChampions(champArray.sort((a, b) => a.name.localeCompare(b.name)));
            } catch (error) {
                console.error("Şampiyonlar yüklenemedi", error);
            }
        };
        fetchChampions();
    }, []);

    // Arama filtresi
    const filteredChamps = useMemo(() => {
        return champions.filter(c => c.name.toLowerCase().includes(search.toLowerCase()));
    }, [search, champions]);

    // Modal kapalıysa hiçbir şey render etme
    if (!isOpen) return null;

    const handleSelect = (id: number) => {
        if (isMulti) {
            setSelectedMulti(prev => 
                prev.includes(id) ? prev.filter(cId => cId !== id) : [...prev, id]
            );
        } else {
            onSelect(id);
            onClose();
        }
    };

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-in fade-in duration-200">
            <div className="bg-[#0B0E14] border border-[#1F242F] rounded-xl w-full max-w-2xl max-h-[80vh] flex flex-col shadow-2xl overflow-hidden">
                
                {/* Modal Header & Search */}
                <div className="p-4 border-b border-[#1F242F] bg-[#151923] flex gap-4 items-center">
                    <div className="flex-1 relative">
                        <input 
                            type="text" 
                            placeholder="Şampiyon ara..." 
                            value={search}
                            onChange={(e) => setSearch(e.target.value)}
                            className="w-full bg-[#0B0E14] border border-[#333C4D] rounded-lg pl-10 pr-4 py-2.5 text-[#E2E8F0] focus:border-[#D4AF37] focus:outline-none transition-colors"
                        />
                        <svg className="w-5 h-5 absolute left-3 top-3 text-[#73829A]" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                        </svg>
                    </div>
                    <button onClick={onClose} className="p-2 text-[#73829A] hover:text-white transition-colors">
                        <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                        </svg>
                    </button>
                </div>

                {/* Champion Grid */}
                <div className="p-4 overflow-y-auto flex-1 grid grid-cols-4 sm:grid-cols-6 md:grid-cols-8 gap-4 custom-scrollbar">
                    {filteredChamps.map(champ => {
                        const isSelected = isMulti && selectedMulti.includes(champ.id);
                        return (
                            <button
                                key={champ.id}
                                onClick={() => handleSelect(champ.id)}
                                className={`relative group flex flex-col items-center gap-2 p-2 rounded-xl transition-all duration-200
                                    ${isSelected ? 'bg-[#D4AF37]/20 border border-[#D4AF37]' : 'hover:bg-[#1F242F] border border-transparent'}`}
                            >
                                <img 
                                    src={champ.image} 
                                    alt={champ.name} 
                                    className={`w-12 h-12 rounded-full object-cover shadow-lg transition-transform duration-200 group-hover:scale-110
                                        ${isSelected ? 'ring-2 ring-[#D4AF37] ring-offset-2 ring-offset-[#0B0E14]' : 'ring-1 ring-[#333C4D]'}`}
                                />
                                <span className={`text-xs truncate w-full text-center ${isSelected ? 'text-[#D4AF37] font-bold' : 'text-[#94A3B8]'}`}>
                                    {champ.name}
                                </span>
                            </button>
                        );
                    })}
                </div>

                {/* Modal Footer (Sadece Çoklu Seçimdeyse Görünür) */}
                {isMulti && (
                    <div className="p-4 border-t border-[#1F242F] bg-[#151923] flex justify-between items-center">
                        <span className="text-sm text-[#73829A]">{selectedMulti.length} Şampiyon Seçildi</span>
                        <button 
                            onClick={() => { onSelect(selectedMulti); onClose(); }}
                            className="bg-[#D4AF37] hover:bg-[#F3D360] text-[#0B0E14] font-bold py-2 px-6 rounded-lg transition-colors shadow-lg active:scale-95"
                        >
                            Havuzu Kaydet
                        </button>
                    </div>
                )}
            </div>
        </div>
    );
};