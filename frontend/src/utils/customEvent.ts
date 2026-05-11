import { CUSTOM_EVENT_KEYS } from '~types/general';

type Listener = (...args: any[]) => void;

class CustomEvent {
  events: { [key: string]: Listener[] | Listener } = {};

  on(
    event: CUSTOM_EVENT_KEYS | string,
    listener: Listener,
    isNoArrayListener?: boolean,
  ) {
    if (!this.events[event] && !isNoArrayListener) {
      this.events[event] = [];
    }

    if (isNoArrayListener) this.events[event.toString()] = listener;
    else {
      if (Array.isArray(this.events[event.toString()]))
        (this.events[event.toString()] as Listener[])?.push(listener);
    }
  }

  off(event: CUSTOM_EVENT_KEYS | string, listener: Listener) {
    const eventLis = this.events[event.toString()];

    if (!Array.isArray(eventLis)) {
      if (event in this.events) delete this.events[event.toString()];
    } else {
      this.events[event.toString()] = eventLis?.filter((l) => l !== listener);
    }
  }

  emit(event: CUSTOM_EVENT_KEYS | string, ...args: unknown[]) {
    const eventLis = this.events[event.toString()];

    if (Array.isArray(eventLis))
      eventLis?.forEach((listener) => listener(...args));
    else if (eventLis) eventLis(...args);
  }
}

export const customEvent = new CustomEvent();
