import { useEffect, useMemo, useState } from 'react';
import { Champion } from '../services/champions';

interface ChampionModalProps {
    isOpen: boolean;
    onClose: () => void;
    onSelect: (id: number | number[]) => void;
    champions: Champion[];
    isMulti?: boolean;
    initialSelected?: number[];
}

export const ChampionModal = ({ isOpen, onClose, onSelect, champions, isMulti = false, initialSelected = [] }: ChampionModalProps) => {
    const [search, setSearch] = useState('');
    const [selectedMulti, setSelectedMulti] = useState<number[]>(initialSelected);

    useEffect(() => {
        if (!isOpen) return;
        setSearch('');
        setSelectedMulti([...initialSelected]);
        const closeOnEscape = (event: KeyboardEvent) => {
            if (event.key === 'Escape') onClose();
        };
        window.addEventListener('keydown', closeOnEscape);
        return () => window.removeEventListener('keydown', closeOnEscape);
    }, [isOpen]);

    const filteredChampions = useMemo(() => {
        const query = search.trim().toLocaleLowerCase('tr-TR');
        if (!query) return champions;
        return champions.filter((champion) => champion.name.toLocaleLowerCase('tr-TR').includes(query));
    }, [search, champions]);

    if (!isOpen) return null;

    const selectChampion = (id: number) => {
        if (isMulti) {
            setSelectedMulti((current) => current.includes(id) ? current.filter((selected) => selected !== id) : [...current, id]);
            return;
        }
        onSelect(id);
        onClose();
    };

    return (
        <div
            role="presentation"
            className="fixed inset-0 z-50 grid place-items-center bg-[#07090D]/85 p-5"
            onMouseDown={(event) => { if (event.target === event.currentTarget) onClose(); }}
        >
            <section role="dialog" aria-modal="true" aria-labelledby="champion-modal-title" className="flex max-h-[82vh] w-full max-w-3xl flex-col overflow-hidden rounded-xl border border-[#303A47] bg-[#151A22]">
                <header className="flex items-center gap-4 border-b border-[#28313D] px-5 py-4">
                    <div className="min-w-0 flex-1">
                        <h2 id="champion-modal-title" className="text-sm font-semibold text-white">{isMulti ? 'Şampiyon havuzu' : 'Şampiyon seç'}</h2>
                        <p className="mt-0.5 text-xs text-[#7F8A9A]">{isMulti ? 'Birden fazla şampiyon seçebilirsin.' : 'Seçim yapıldığında otomatik kaydedilir.'}</p>
                    </div>
                    <button type="button" onClick={onClose} aria-label="Pencereyi kapat" className="grid h-8 w-8 place-items-center rounded-md text-xl leading-none text-[#7F8A9A] hover:bg-[#222A34] hover:text-white">×</button>
                </header>

                <div className="border-b border-[#28313D] px-5 py-3">
                    <label className="relative block">
                        <span className="sr-only">Şampiyon ara</span>
                        <svg className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 fill-none stroke-[#6F7A8A] stroke-[1.6]" viewBox="0 0 20 20" aria-hidden="true"><circle cx="8.5" cy="8.5" r="5.5" /><path d="m13 13 4 4" /></svg>
                        <input
                            autoFocus
                            type="search"
                            value={search}
                            onChange={(event) => setSearch(event.target.value)}
                            placeholder="İsme göre ara"
                            className="h-10 w-full rounded-lg border border-[#303A47] bg-[#10151C] pl-10 pr-4 text-sm text-[#E8EBF0] placeholder:text-[#626E7E]"
                        />
                    </label>
                </div>

                <div className="grid min-h-40 flex-1 grid-cols-4 gap-2 overflow-y-auto p-4 sm:grid-cols-6 md:grid-cols-8">
                    {filteredChampions.map((champion) => {
                        const selected = isMulti && selectedMulti.includes(champion.id);
                        return (
                            <button
                                type="button"
                                key={champion.id}
                                aria-pressed={selected}
                                onClick={() => selectChampion(champion.id)}
                                className={`flex min-w-0 flex-col items-center gap-2 rounded-lg border p-2 text-center transition-colors ${selected ? 'border-[#C8A24A] bg-[#2A2417]' : 'border-transparent hover:border-[#384452] hover:bg-[#1D242D]'}`}
                            >
                                <img src={champion.image} alt="" className="h-11 w-11 rounded-lg object-cover" />
                                <span className={`w-full truncate text-[11px] ${selected ? 'font-semibold text-[#D7B967]' : 'text-[#9AA4B2]'}`}>{champion.name}</span>
                            </button>
                        );
                    })}
                    {filteredChampions.length === 0 && <p className="col-span-full py-12 text-center text-sm text-[#7F8A9A]">Eşleşen şampiyon bulunamadı.</p>}
                </div>

                {isMulti && (
                    <footer className="flex items-center justify-between gap-4 border-t border-[#28313D] px-5 py-4">
                        <span className="text-xs text-[#8994A4]">{selectedMulti.length} şampiyon seçildi</span>
                        <button type="button" onClick={() => { onSelect(selectedMulti); onClose(); }} className="flat-button border-[#C8A24A] bg-[#C8A24A] text-[#111318] hover:bg-[#D5B55E]">
                            Havuzu kaydet
                        </button>
                    </footer>
                )}
            </section>
        </div>
    );
};
