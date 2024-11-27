export interface CustomWindow extends Window {
  electron: any
}

declare let window: CustomWindow


declare global {
  interface Window {
    electron
  }
}

export interface Config {
  innerHost?: string
  frontHost?: string
}