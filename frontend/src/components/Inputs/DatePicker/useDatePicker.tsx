import { useMemo } from 'react';

import { useRefAndState } from '~hooks/useRefAndState';

const checkDate = (date: Date, minDate: Date, maxDate: Date) => {
  return !(date < minDate || date > maxDate);
};

type DateTimePickerProps = {
  minDate?: Date;
  maxDate?: Date;
  value?: Date;
  onChange?: (date: Date) => void;
};

export const useDateTimePicker = ({
  minDate = new Date(1970, 0, 1),
  maxDate = new Date(2100, 11, 31, 23, 59, 59),
  value,
  onChange,
}: DateTimePickerProps) => {
  const [getSelected, setSelected, selected] = useRefAndState<Date>(
    value ?? new Date(),
  );

  const update = (date: Date) => {
    let newDate = date;

    if (minDate && newDate < minDate) {
      newDate = minDate;
    }

    if (maxDate && newDate > maxDate) {
      newDate = maxDate;
    }

    setSelected(newDate);
    onChange?.(newDate);
  };

  // --- Options generators ---
  const years = useMemo(() => {
    const arr: number[] = [];

    for (let y = minDate.getFullYear(); y <= maxDate.getFullYear(); y++) {
      arr.push(y);
    }

    return arr;
  }, [minDate, maxDate]);

  const months = useMemo(() => {
    const start =
      selected.getFullYear() === minDate.getFullYear() ? minDate.getMonth() : 0;
    const end =
      selected.getFullYear() === maxDate.getFullYear()
        ? maxDate.getMonth()
        : 11;

    return Array.from({ length: end - start + 1 }, (_, i) => start + i);
  }, [selected, minDate, maxDate]);

  const days = useMemo(() => {
    const year = selected.getFullYear();
    const month = selected.getMonth();

    const lastDay = new Date(year, month + 1, 0);

    const start =
      year === minDate.getFullYear() && month === minDate.getMonth()
        ? minDate.getDate()
        : 1;
    const end =
      year === maxDate.getFullYear() && month === maxDate.getMonth()
        ? maxDate.getDate()
        : lastDay.getDate();

    return Array.from({ length: end - start + 1 }, (_, i) => start + i);
  }, [selected, minDate, maxDate]);

  const hours = useMemo(() => {
    const year = selected.getFullYear();
    const month = selected.getMonth();
    const day = selected.getDate();

    const isMinDay =
      year === minDate.getFullYear() &&
      month === minDate.getMonth() &&
      day === minDate.getDate();

    const isMaxDay =
      year === maxDate.getFullYear() &&
      month === maxDate.getMonth() &&
      day === maxDate.getDate();

    const start = isMinDay ? minDate.getHours() : 0;
    const end = isMaxDay ? maxDate.getHours() : 23;

    return Array.from({ length: end - start + 1 }, (_, i) => start + i);
  }, [selected, minDate, maxDate]);

  const minutes = useMemo(() => {
    const isMinHour =
      selected.toDateString() === minDate.toDateString() &&
      selected.getHours() === minDate.getHours();

    const isMaxHour =
      selected.toDateString() === maxDate.toDateString() &&
      selected.getHours() === maxDate.getHours();

    const start = isMinHour ? minDate.getMinutes() : 0;
    const end = isMaxHour ? maxDate.getMinutes() : 59;

    return Array.from({ length: end - start + 1 }, (_, i) => start + i);
  }, [selected, minDate, maxDate]);

  const seconds = useMemo(() => {
    const isMinMinute =
      selected.toDateString() === minDate.toDateString() &&
      selected.getHours() === minDate.getHours() &&
      selected.getMinutes() === minDate.getMinutes();

    const isMaxMinute =
      selected.toDateString() === maxDate.toDateString() &&
      selected.getHours() === maxDate.getHours() &&
      selected.getMinutes() === maxDate.getMinutes();

    const start = isMinMinute ? minDate.getSeconds() : 0;
    const end = isMaxMinute ? maxDate.getSeconds() : 59;

    return Array.from({ length: end - start + 1 }, (_, i) => start + i);
  }, [selected, minDate, maxDate]);

  type PartialDate = {
    year?: number;
    month?: number;
    day?: number;
    hour?: number;
    minute?: number;
    second?: number;
  };

  const setDateTime = ({
    year,
    month,
    day,
    hour,
    minute,
    second,
  }: PartialDate) => {
    update(
      new Date(
        year ?? selected.getFullYear(),
        month ?? selected.getMonth(),
        day ?? selected.getDate(),
        hour ?? selected.getHours(),
        minute ?? selected.getMinutes(),
        second ?? selected.getSeconds(),
      ),
    );
  };

  return {
    selected,
    years,
    months,
    days,
    hours,
    minutes,
    seconds,
    setDateTime,
    checkDate,
    getSelected,
  };
};
