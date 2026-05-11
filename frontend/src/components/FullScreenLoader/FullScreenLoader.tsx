import { DuckLoad } from '~icons/JsxSvg/DuckLoad';
import { LoadIcon } from '~icons/JsxSvg/LoadIcon';
import colors from '~styles/colors';

export const FullScreenLoader = ({ progress }: { progress: number }) => {
  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        flexDirection: 'column',
        gap: '2rem',
        width: '100vw',
        height: '100vh',
        position: 'fixed',
        top: 0,
        left: 0,
        zIndex: 100,
        overflow: 'hidden',
        background: colors.backgroundDark,
        transition: `opacity 300ms ease-out`,
      }}
    >
      <DuckLoad height={200} width={200} />

      <div
        style={{
          display: 'flex',
          flexDirection: 'row',
          alignItems: 'center',
          justifyContent: 'center',
          width: '200px',
          height: '56px',
          background: 'transparent',
          borderRadius: '5px',
          color: colors.textSecondary,
        }}
      >
        <div style={{ width: '150px', height: '34px', position: 'relative' }}>
          {/* Progress mask */}
          <div
            style={{
              width: `${progress}%`,
              height: '100%',
              color: colors.primary,
              position: 'absolute',
              top: 0,
              left: 0,
              overflow: 'hidden',
              transition: 'width 100ms ease',
            }}
          >
            <LoadIcon width={150} height={34} />
          </div>
          <LoadIcon width={150} height={34} />
        </div>
      </div>
    </div>
  );
};
