interface Props {
  onClick?: () => void;
  className?: string;
}

export const Chevron = ({ onClick, className }: Props) => {
  return (
    <svg
      aria-label={'Toggle'}
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      onClick={onClick}
      className={className}
      viewBox="0 0 24 24"
      fill="none"
    >
      <path
        d="M4.36328 8.72754L11.9996 16.3639L19.636 8.72754"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
};
