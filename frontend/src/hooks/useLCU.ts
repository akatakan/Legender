import { useCallback, useEffect, useRef, useState } from 'react';
import { fetchLCUConnection } from '../services/lcu';
import { main } from '../../wailsjs/go/models';

export const useLCU = () => {
    const [lcuData, setLcuData] = useState<main.LCUResponse | null>(null);
    const [isLoading, setIsLoading] = useState(false);
	const [error, setError] = useState<string | null>(null);
	const requestInFlight = useRef(false);
	const mounted = useRef(true);

    const refreshConnection = useCallback(async (showLoading: boolean) => {
		if (requestInFlight.current) return;
		requestInFlight.current = true;
		if (showLoading && mounted.current) setIsLoading(true);
        try {
            const data = await fetchLCUConnection();
			if (mounted.current) setLcuData(data);
			if (mounted.current) setError(null);
        } catch (error) {
            console.error("LCU verileri çekilirken hata oluştu:", error);
			if (mounted.current) {
				setLcuData(main.LCUResponse.createFrom({ isActive: false, port: '' }));
				setError('League Client bağlantısı kontrol edilemedi. Ayrıntılar uygulama loguna yazıldı.');
			}
        } finally {
			requestInFlight.current = false;
			if (showLoading && mounted.current) setIsLoading(false);
        }
	}, []);

	useEffect(() => {
		mounted.current = true;
		void refreshConnection(false);
		const interval = window.setInterval(() => void refreshConnection(false), 5000);
		return () => {
			mounted.current = false;
			window.clearInterval(interval);
		};
	}, [refreshConnection]);

	const checkConnection = useCallback(async () => {
		await refreshConnection(true);
	}, [refreshConnection]);
	const clearError = useCallback(() => setError(null), []);

    return { lcuData, isLoading, error, checkConnection, clearError };
};
