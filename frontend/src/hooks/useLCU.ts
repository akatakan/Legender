import { useState } from 'react';
import { fetchLCUConnection } from '../services/lcu';
import { main } from '../../wailsjs/go/models';

export const useLCU = () => {
    const [lcuData, setLcuData] = useState<main.LCUResponse | null>(null);
    const [isLoading, setIsLoading] = useState(false);

    const checkConnection = async () => {
        setIsLoading(true);
        try {
            const data = await fetchLCUConnection();
            setLcuData(data);
        } catch (error) {
            console.error("LCU verileri çekilirken hata oluştu:", error);
        } finally {
            setIsLoading(false);
        }
    };

    return { lcuData, isLoading, checkConnection };
};