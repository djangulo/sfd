import _uniqBy from "lodash/uniqBy";
import _uniq from "lodash/uniq";

import ItemsAPI from "../services/item";

// state structure
// export interface ItemsState {
//   items[];
//   bids;
//   subscribed[];
//   selectedItem | null;
//   searchQuery | null;
//   errors: object | null;
//   isLoading: boolean;
//   showSubscribed: boolean;
//   showClosed: boolean;
//   lastId | null;
// }

export const initialState = {
  items: [],
  bids: [],
  selectedItem: null,
  searchQuery: "",
  errors: null,
  isLoading: false,
  showSubscribed: false,
  showClosed: false,
  lastID: null,
  lastCreatedAt: null,
};

// action types
const APPEND_BID = "APPEND_BID";
const APPEND_ITEM = "APPEND_ITEM";
const INVALIDATE_BID = "INVALIDATE_BID";
const SET_BIDS = "SET_BIDS";
const SET_ERROR = "SET_ERROR";
const SET_FORM_ERRORS = "SET_FORM_ERRORS";
const SET_ITEMS = "SET_ITEMS";
const SET_LOADING = "SET_LOADING";
const SET_SELECTED_ITEM = "SET_SELECTED_ITEM";
const SET_SUBSCRIBED = "SET_SUBSCRIBED";
const UNSET_SUBSCRIBED = "UNSET_SUBSCRIBED";
const SET_WINNING_BID = "SET_WINNING_BID";

// action creators

export const appendBid = (bid) => ({
  type: APPEND_BID,
  bid,
});
export const appendItem = (item) => ({
  type: APPEND_ITEM,
  item,
});
export const invalidateBid = (bid) => ({
  type: INVALIDATE_BID,
  bid,
});
export const setBids = (bids) => ({
  type: SET_BIDS,
  bids,
});
export const setError = (error) => ({
  type: SET_ERROR,
  error,
});
/**
 *
 * @param errors key value object of form errors, where keys map to form-fields
 */
export const setFormErrors = (errors) => ({
  type: SET_FORM_ERRORS,
  errors,
});
export const setItems = (items) => ({
  type: SET_ITEMS,
  items,
});
export const setLoading = (isLoading) => ({
  type: SET_LOADING,
  isLoading,
});

export const setSelectedItem = (item) => ({
  type: SET_SELECTED_ITEM,
  item,
});

export const setSubscribed = (itemId) => ({
  type: SET_SUBSCRIBED,
  itemId,
});

export const unsetSubscribed = (itemId) => ({
  type: UNSET_SUBSCRIBED,
  itemId,
});

export const setWinningBid = (bid) => ({
  type: SET_WINNING_BID,
  bid,
});

// export type ItemActions =
//   | ActionType<typeof appendBid>
//   | ActionType<typeof appendItem>
//   | ActionType<typeof invalidateBid>
//   | ActionType<typeof setBids>
//   | ActionType<typeof setError>
//   | ActionType<typeof setFormErrors>
//   | ActionType<typeof setItems>
//   | ActionType<typeof setLoading>
//   | ActionType<typeof setSelectedItem>
//   | ActionType<typeof setSubscribed>
//   | ActionType<typeof unsetSubscribed>;

// async workflows

export const selectItem = (itemID) => async (dispatch) => {
  dispatch(setLoading(true));
  ItemsAPI.retrieve(itemID)
    .then((response) => {
      dispatch(setLoading(false));
      dispatch(setSelectedItem(response.data));
    })
    .catch((error) => {
      dispatch(setLoading(false));
      dispatch(setError(error));
    });
};

export const fetchItems = (opts) => async (dispatch) => {
  dispatch(setLoading(true));
  ItemsAPI.list({ ...opts })
    .then((response) => {
      dispatch(setLoading(false));
      dispatch(setItems(response.data));
    })
    .catch((error) => {
      dispatch(setLoading(false));
    });
};

/**
 *
 * @param errors comma-separated string of key:value pairs
 */
const parseErrors = (errors) => {
  const obj = {};
  errors
    .split(",")
    .map((e) => e.split(":"))
    .map((pair) => (obj[pair[0]] = pair[1]));
  return obj;
};

export const createItem = (item) => async (dispatch) => {
  dispatch(setLoading(true));
  ItemsAPI.create(item)
    .then((response) => {
      dispatch(setLoading(false));
      dispatch(appendItem(response.data));
    })
    .catch((errors) => {
      dispatch(setLoading(false));
      dispatch(setFormErrors(parseErrors(errors)));
    });
};

export const getItemBids = (itemId, opts) => async (dispatch) => {
  dispatch(setLoading(true));
  ItemsAPI.itemBids(itemId, opts)
    .then((response) => {
      dispatch(setLoading(false));
      dispatch(setBids(response.data));
    })
    .catch((error) => {
      dispatch(setLoading(false));
      dispatch(setError(error));
    });
};

export const getItemWinningBid = (itemId) => async (dispatch) => {
  dispatch(setLoading(true));
  ItemsAPI.itemWinningBid(itemId)
    .then((response) => {
      dispatch(setLoading(false));
      dispatch(setWinningBid(response.data));
    })
    .catch((error) => {
      dispatch(setError(error));
    });
};

export const getUserBids = (userId, opts) => async (dispatch) => {
  dispatch(setLoading(true));
  ItemsAPI.itemBids(userId, opts)
    .then((response) => {
      dispatch(setLoading(false));
      dispatch(setBids(response.data));
    })
    .catch((error) => {
      dispatch(setLoading(false));
      dispatch(setError(error));
    });
};

export const subscribe = (userId, itemId) => async (dispatch) => {
  dispatch(setLoading(true));
  ItemsAPI.subscribe(userId, itemId)
    .then((response) => {
      dispatch(setLoading(false));
      dispatch(setSubscribed(response.data));
    })
    .catch((error) => {
      dispatch(setLoading(false));
      dispatch(setError(error));
    });
};

export const unsubscribe = (userId, itemId) => async (dispatch) => {
  dispatch(setLoading(true));
  ItemsAPI.unsubscribe(userId, itemId)
    .then((response) => {
      dispatch(setLoading(false));
      dispatch(unsetSubscribed(response.data));
    })
    .catch((error) => {
      dispatch(setLoading(false));
      dispatch(setError(error));
    });
};

// selectors

export const getItems = ({ items: { items } }) => items || [];
export const selectedItem = ({ items: { selectedItem } }) =>
  selectedItem || null;
export const itemsError = ({ items: { errors } }) => errors || "";
export const itemsIsLoading = ({ items: { isLoading } }) => isLoading || null;
export const itemsLastID = ({ items: { lastID } }) => lastID || null;
export const itemsLastCreatedAt = ({ items: { lastCreatedAt } }) =>
  lastCreatedAt || null;
export const itemIDFromSlug = ({ items: { items } }) => (slug) => {
  const it = items.find((i) => i.slug === slug);
  if (it && it.id) return it.id;
  return "";
};

// reducer
const reducer = (state = initialState, action) => {
  switch (action.type) {
    case APPEND_BID: {
      return {
        ...state,
        isLoading: false,
        bids: [...state.bids, action.bid],
      };
    }
    case APPEND_ITEM: {
      return {
        ...state,
        isLoading: false,
        selectedItem: action.item,
        items: [...state.items, action.item],
      };
    }
    case INVALIDATE_BID: {
      return {
        ...state,
        bids: [
          ...state.bids.filter((bid) => bid.id !== action.bid.id),
          action.bid,
        ],
      };
    }
    case SET_BIDS: {
      return {
        ...state,
        isLoading: false,
        bids: [...action.bids],
      };
    }
    case SET_ERROR: {
      return { ...state, errors: { error: action.error } };
    }
    case SET_FORM_ERRORS: {
      return {
        ...state,
        isLoading: false,
        errors: { ...action.errors },
      };
    }
    case SET_ITEMS: {
      return {
        ...state,
        isLoading: false,
        items: _uniqBy([...state.items, ...action.items], "id"),
        lastId: action.items.length
          ? action.items[action.items.length - 1].id
          : state.items[state.items.length - 1].id,
      };
    }
    case SET_LOADING: {
      return { ...state, isLoading: action.isLoading };
    }
    case SET_SELECTED_ITEM: {
      return { ...state, selectedItem: action.item };
    }
    case SET_SUBSCRIBED: {
      return {
        ...state,
        isLoading: false,
        subscribed: _uniq([...state.subscribed, action.itemId]),
      };
    }
    case UNSET_SUBSCRIBED: {
      return {
        ...state,
        isLoading: false,
        subscribed: [...state.subscribed.filter((v) => v !== action.itemId)],
      };
    }
    case SET_WINNING_BID: {
      return {
        ...state,
        selectedItem: { ...state.selectedItem, winning_bid: action.bid },
      };
    }
    default: {
      return { ...state };
    }
  }
};

export default reducer;
