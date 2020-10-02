export const initialState = {
  currentRouteKey: "items.home.path",
};

const SET_ROUTE_KEY = "SET_ROUTE_KEY";

export const setRouteKey = (key) => ({
  type: SET_ROUTE_KEY,
  key,
});

export const i18nGetKey = ({ i18n: { currentRouteKey } }) =>
  currentRouteKey || "";

const reducer = (state = initialState, action) => {
  switch (action.type) {
    case SET_ROUTE_KEY: {
      return {
        ...state,
        currentRouteKey: action.key,
      };
    }
    default: {
      return { ...state };
    }
  }
};

export default reducer;
