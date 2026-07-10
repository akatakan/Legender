import { ImgHTMLAttributes, useEffect, useMemo, useState } from 'react';

interface FallbackImageProps extends Omit<ImgHTMLAttributes<HTMLImageElement>, 'src'> {
    sources: string[];
}

export const FallbackImage = ({ sources, ...props }: FallbackImageProps) => {
    const sourceKey = useMemo(() => sources.join('|'), [sources]);
    const [sourceIndex, setSourceIndex] = useState(0);

    useEffect(() => setSourceIndex(0), [sourceKey]);

    return (
        <img
            {...props}
            src={sources[sourceIndex] ?? ''}
            onError={() => setSourceIndex((current) => Math.min(current + 1, sources.length - 1))}
        />
    );
};
