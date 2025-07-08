import i18n from "i18next";
import { initReactI18next } from "react-i18next";
import LanguageDetector from 'i18next-browser-languagedetector';

import zhCN from './locales/zh-CN/translation.json'
import enUS from './locales/en-US/translation.json'
import config from '../config'

const fallbackLng = config.defaultLang || 'enUS'

let resources: Record<string, Record<string, any>> = {
  zhCN: {
    translation: zhCN
  },
  enUS: {
    translation: enUS
  },
};

if (config.brandName === 'truewatch') {
  resources = {
    enUS: {
      translation: enUS
    },
  }
}

i18n.use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources,
    fallbackLng,
    detection: {
      order: ['sessionStorage', 'localStorage'],
      caches: ['sessionStorage', 'localStorage', 'cookie'],
    },
  })

let inited = false

export function initLanguage(lang: string) {
  if (!inited) {
    lang = lang === "zh" ? "zhCN" : "enUS"
    i18n.changeLanguage(lang)
    inited = true
  }
}

export function toggleLanguage() {
  const lang = i18n.language === "zhCN" ? "enUS" : "zhCN"
  i18n.changeLanguage(lang)
}

export default i18n