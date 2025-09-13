import { Middleware } from 'redux';
import { PayloadAction } from '@reduxjs/toolkit';
import { ActionComponent } from '../../api/model';
import { SocketMessage, ActionMessage } from '../../api/socket';
import { ACTION_COMPONENT } from '../types/actions';
import { onlineSet } from '../actions/online';
import { reset, componentAdd, componentRemove, attributeChange } from '../actions/registry';
import { modelInit } from '../actions/model';

const PING_INTERVAL = 2000;          // Send ping every 2s
const IDLE_TIMEOUT = 5000;           // If no messages for 5s → reconnect
const BASE_RECONNECT_DELAY = 1000;   // Start retry after 1s
const MAX_RECONNECT_DELAY = 10000;   // Cap retries at 10s

export const socketMiddleware: Middleware = (store) => (next) => {
  const url = location.origin.replace(/^http/, 'ws') + '/ws';
  const socket = new ReconnectingWebSocket(url);

  socket.onopen = () => {
    next(onlineSet(true));
  };

  socket.onclose = () => next(onlineSet(false));

  socket.onmessage = (type, data) => {
    switch (type) {
      case 'state':
        next(reset(data));
        break;

      case 'add':
        next(componentAdd(data));
        break;

      case 'remove':
        next(componentRemove(data));
        break;

      case 'change':
        next(attributeChange(data));
        break;

      case 'modelHash':
        next(modelInit(data) as any); // TODO: proper cast: AppThunkAction => AnyAction
        break;
    }
  };

  return (action) => {
    if (action.type === ACTION_COMPONENT) {
      const typedAction = action as PayloadAction<ActionComponent>;
      socket.send('action', typedAction.payload as ActionMessage);
    }

    return next(action);
  };
};

type Timer = ReturnType<typeof setTimeout> | ReturnType<typeof setInterval>;
type SocketMessageType = SocketMessage['type'];
class ReconnectingWebSocket {
  private url: string;
  private ws: WebSocket | null = null;
  private pingInterval: Timer | null = null;
  private idleTimeout: Timer | null = null;
  private reconnectDelay: number = BASE_RECONNECT_DELAY;

  // User callbacks
  public onopen: () => void = () => {};
  public onclose: () => void = () => {};
  public onmessage: (type: SocketMessageType, data: any) => void = () => {};

  constructor(url: string) {
    this.url = url;
    this.connect();
  }

  private connect(): void {
    this.ws = new WebSocket(this.url);

    this.ws.onopen = () => {
      console.log("Connected to", this.url);

      // Reset backoff
      this.reconnectDelay = BASE_RECONNECT_DELAY;

      // Start ping loop
      this.pingInterval = setInterval(() => {
        if (this.ws?.readyState === WebSocket.OPEN) {
          this.send("ping", null);
        }
      }, PING_INTERVAL);

      this.resetIdleTimer();
      this.onopen();
    };

    this.ws.onmessage = (event: MessageEvent) => {
      this.resetIdleTimer();

      try {
        const { type, data } = JSON.parse(event.data) as SocketMessage;
        this.onmessage(type, data);
      } catch (error) {
        console.error("Error handling WebSocket message:", error);
      }
    };

    this.ws.onclose = () => {
      console.warn("Connection closed, retrying in", this.reconnectDelay, "ms");
      this.cleanup();
      setTimeout(() => this.connect(), this.reconnectDelay);

      // Exponential backoff
      this.reconnectDelay = Math.min(this.reconnectDelay * 2, MAX_RECONNECT_DELAY);

      this.onclose();
    };

    this.ws.onerror = (err: Event) => {
      console.error("WebSocket error:", err);
      this.ws?.close(); // triggers onclose
    };
  }

  private resetIdleTimer(): void {
    if (this.idleTimeout) clearTimeout(this.idleTimeout);
    this.idleTimeout = setTimeout(() => {
      console.warn("Idle timeout, forcing reconnect...");
      this.ws?.close(); // triggers onclose
    }, IDLE_TIMEOUT);
  }

  private cleanup(): void {
    if (this.pingInterval) clearInterval(this.pingInterval);
    if (this.idleTimeout) clearTimeout(this.idleTimeout);
  }

  public send(type: SocketMessageType, data: any) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      const message: SocketMessage = { type, data };
      this.ws.send(JSON.stringify(message));
    } else {
      console.warn("Cannot send, socket not open");
    }
  }
}
