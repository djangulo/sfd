import { combineReducers } from "redux";

import items from "./itemsDuck";
import i18n from "./i18nDuck";
// import pages from './pagesDuck';
// import search from './searchDuck';

export * from "./itemsDuck";
export * from "./i18nDuck";

export const rootReducer = combineReducers({
  items,
  i18n,
  // search,
});
