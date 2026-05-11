export interface TimeLeftType {
  days: number;
  hours: number;
  minutes: number;
  seconds: number;
}

export function getTimeLeft(endDate: number): TimeLeftType {
  const now = new Date().getTime();
  const timeDiff = endDate - now;

  if (timeDiff <= 0) {
    return { days: 0, hours: 0, minutes: 0, seconds: 0 };
  }

  const seconds = Math.floor((timeDiff / 1000) % 60);
  const minutes = Math.floor((timeDiff / 1000 / 60) % 60);
  const hours = Math.floor((timeDiff / (1000 * 60 * 60)) % 24);
  const days = Math.floor(timeDiff / (1000 * 60 * 60 * 24));

  return { days, hours, minutes, seconds };
}

export const getTimeLeftString = (
  timeLeft: TimeLeftType,
): {
  key: 'times-up' | 'time';
  value: string;
} => {
  const { days, hours, minutes, seconds } = timeLeft;

  if (days === 0 && hours === 0 && minutes === 0 && seconds === 0) {
    return {
      value: "Time's up!",
      key: 'times-up',
    };
  }

  const daysStr = `${days}d`;
  const hoursStr = `${hours.toString().length === 1 ? `0${hours}` : hours}h`;
  const minutesStr = `${minutes.toString().length === 1 ? `0${minutes}` : minutes}m`;
  const secondsStr = `${seconds.toString().length === 1 ? `0${seconds}` : seconds}s`;

  const timeValue =
    (days > 0 && `${daysStr}:${hoursStr}:${minutesStr}`) ||
    (hours > 0 && `${hoursStr}:${minutesStr}:${secondsStr}`) ||
    (minutes > 0 && `${minutesStr}:${secondsStr}`) ||
    (seconds >= 0 && `${secondsStr} `) ||
    '';

  return {
    value: `${timeValue}`,
    key: 'time',
  };
};
