import { Window, Model as NetModel, DefaultWindow, Control } from '../../server-api/model';

export { NetModel };
export { Window, Control, ControlDisplay, ControlDisplayMapItem, ControlText, ControlTextContextItem, Action, ActionComponent, ActionWindow, Resource } from '../../server-api/model';

export const MODEL_SET = 'model/set';

export interface Model {
  defaultWindow: DefaultWindow;
  windows: { [id: string]: Window };
  controls: { [id: string]: Control };
}