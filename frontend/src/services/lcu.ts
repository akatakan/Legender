import { GetLCUData, ToggleAutoAccept, GetSettings, SaveAutoPickSettings } from '../../wailsjs/go/main/App';
import { main,config } from '../../wailsjs/go/models';

export const fetchLCUConnection = async (): Promise<main.LCUResponse> => {
    return await GetLCUData();
};

export const setAutoAcceptState = async (enabled: boolean): Promise<void> => {
    await ToggleAutoAccept(enabled);
};

export const fetchSettings = async (): Promise<config.Settings> => {
    return await GetSettings();
};

export const updateAutoPickSettings = async (cfg: config.AutoPickSettings): Promise<void> => {
    await SaveAutoPickSettings(cfg);
};