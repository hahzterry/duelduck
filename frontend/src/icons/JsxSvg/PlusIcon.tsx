import colors from '~styles/colors';
import { ColorType } from '~types/general';

export const PlusIcon = ({ color }: { color?: ColorType }) => {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="18"
      height="18"
      viewBox="0 0 18 18"
      fill="none"
    >
      <path
        d="M3 9H15M9 3V15"
        stroke={color ? colors[color] : 'currentColor'}
        strokeWidth="1.5"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
};
