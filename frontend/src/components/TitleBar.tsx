import { Quit, WindowMinimise, WindowToggleMaximise } from '../../wailsjs/runtime/runtime';

interface TitleBarProps {
    connected: boolean;
}

export const TitleBar = ({ connected }: TitleBarProps) => (
    <div className="window-drag flex h-11 shrink-0 select-none items-center border-b border-[#252D38] bg-[#11161D]">
        <div className="flex min-w-0 flex-1 items-center gap-3 px-4" onDoubleClick={WindowToggleMaximise}>
            <div className="grid h-6 w-6 place-items-center rounded-md bg-[#C8A24A] text-[11px] font-black text-[#111318]">L</div>
            <span className="truncate text-sm font-semibold tracking-wide text-[#E7EAF0]">Legender</span>
            <span className="h-4 w-px bg-[#2B3440]" />
            <span className="flex items-center gap-2 text-xs text-[#8B96A7]">
                <span className={`h-1.5 w-1.5 rounded-full ${connected ? 'bg-[#43C98B]' : 'bg-[#657080]'}`} />
                {connected ? 'League Client bağlı' : 'İstemci bekleniyor'}
            </span>
        </div>

        <div className="window-no-drag flex h-full items-stretch" aria-label="Pencere kontrolleri">
            <button type="button" title="Küçült" aria-label="Pencereyi küçült" onClick={WindowMinimise} className="window-control">
                <svg viewBox="0 0 12 12" aria-hidden="true"><path d="M2 8.5h8" /></svg>
            </button>
            <button type="button" title="Büyüt veya geri yükle" aria-label="Pencereyi büyüt veya geri yükle" onClick={WindowToggleMaximise} className="window-control">
                <svg viewBox="0 0 12 12" aria-hidden="true"><rect x="2.25" y="2.25" width="7.5" height="7.5" rx="0.5" /></svg>
            </button>
            <button type="button" title="Kapat" aria-label="Uygulamayı kapat" onClick={Quit} className="window-control window-control-close">
                <svg viewBox="0 0 12 12" aria-hidden="true"><path d="m2.5 2.5 7 7m0-7-7 7" /></svg>
            </button>
        </div>
    </div>
);
