import i18n from "i18next";
import ChainedBackend from "i18next-chained-backend";
import HTTPBackend from "i18next-http-backend";
import LocalStorageBackend from "i18next-localstorage-backend";
import LanguageDetector from "i18next-browser-languagedetector";
import { initReactI18next } from "react-i18next";

i18n
  .use(ChainedBackend)
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    initImmediate: false,
    backend: {
      backends: [LocalStorageBackend, HTTPBackend],
      backendOptions: [
        {
          // LocalStorageOptions
          prefix: "sfd_i18next_res_",
          expirationTime: 60000, // 7*24*60*60*1000=604800000 miliseconds (7days),
          versions: {},
          store: window.localStorage,
        },
        {
          // FileSystemBackend Options
          loadPath: "/locales/{{lng}}/{{ns}}.json",
          addPath: "/locales/{{lng}}/{{ns}}.missing.json",
        },
      ],
    },
    lng: "en-US",
    fallbackLng: "en-US",
    saveMissing: true,
    supportedLngs: ["en-US", "es-DO", "en", "es"],
    load: ["en-US", "es-DO"],
    preload: ["en-US", "es-DO"],
    ns: ["router", "languages", "default"],
    returnObjects: true,
    defaultNS: "router",
    saveMissingTo: "fallback",
    debug: true,
    interpolation: {
      escapeValue: false,
    },
  });

export default i18n;
