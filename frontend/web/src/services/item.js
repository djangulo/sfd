import axios from "axios";
// import { Item } from "../models/Item";
// import { Bid } from "../models/Bid";
// import { Image } from "../models/Image";

import config from "./config";
// import { invalidateBid, unsubscribe, unsetSubscribed } from "../data/itemsDuck";

// // export type filterOpts = Nullable<Opts>;
// /**
//  * @root root URL for the service
//  * @itemsEndopint items endpoint path
//  * @usersEndpoint users endpoint path
//  */
// interface IService {
//   root;
//   itemsEndpoint: string;
//   usersEndpoint: string;
//   list(opts: filterOpts | null): Promise<AxiosResponse<Item[]>>;
//   retrieve(id: string): Promise<AxiosResponse<Item>>;
//   create(item: Item): Promise<AxiosResponse<Item>>;
//   itemBids(
//     itemId: string,
//     opts: filterOpts | null
//   ): Promise<AxiosResponse<Bid[]>>;
//   itemImages(
//     itemId: string,
//     opts: filterOpts | null
//   ): Promise<AxiosResponse<Image[]>>;
//   userBids(
//     userId: string,
//     opts: filterOpts | null
//   ): Promise<AxiosResponse<Bid[]>>;
//   subscribe(userId: string, itemId: string): Promise<AxiosResponse<string>>;
//   unsubscribe(userId: string, itemId: string): Promise<AxiosResponse<string>>;
//   placeBid(
//     userId: string,
//     itemId: string,
//     amount: number
//   ): Promise<AxiosResponse<Bid>>;
//   invalidateBid(userId: string, bidId: string): Promise<AxiosResponse<Bid>>;
// }

const resolveOpts = (opts) => {
  const params = { limit: 10 };
  if (opts && opts.limit) {
    params.limit = true;
  }
  if (opts && opts.lastCreatedAt) {
    params.lastCreatedAt = true;
  }
  if (opts && opts.lastID && opts.lastID !== "") {
    params["last-id"] = opts.lastID;
  }
  if (opts && opts.showInvalid) {
    params["show-invalid"] = opts.showInvalid;
  }
  if (opts && opts.showClosed) {
    params["show-closed"] = true;
  }
  if (opts && opts.subscribedByID !== "") {
    params["subscribed-by-id"] = opts.subscribedByID;
  }
  return params;
};

function list(opts = null) {
  const params = resolveOpts(opts);
  return axios.get(`${config.itemsEndpoint}`, { params });
}
function retrieve(id) {
  return axios.get(`${config.itemsEndpoint}/${id}`);
}
function create(item) {
  return axios.post(`${config.itemsEndpoint}`, { ...item });
}
function itemBids(itemId, opts = null) {
  const params = resolveOpts(opts);
  return axios.get(`${config.itemsEndpoint}/${itemId}/bids`, {
    params,
  });
}
function itemWinningBid(itemId) {
  return axios.get(`${config.itemsEndpoint}/${itemId}/bids/winning`);
}
function itemImages(itemId, opts = null) {
  const params = resolveOpts(opts);
  return axios.get(`${config.itemsEndpoint}/images`, {
    params,
  });
}
function userBids(userId, opts = null) {
  const params = resolveOpts(opts);
  return axios.get(`${config.usersEndpoint}/${userId}/bids`, {
    params,
  });
}
function subscribe(userId, itemId) {
  return axios.post(`${config.itemsEndpoint}/${itemId}/subscribe`, {
    "user-id": userId,
  });
}
function unsubscribe(userId, itemId) {
  return axios.post(`${config.itemsEndpoint}/${itemId}/unsubscribe`, {
    "user-id": userId,
  });
}
function placeBid(userId, itemId, amount) {
  return axios.post(`${config.itemsEndpoint}/${itemId}/bids`, {
    amount,
    "user-id": userId,
  });
}
function invalidateBid(userId, bidId) {
  return axios.post(`${config.bidsEndpoint}/${bidId}/invalidate`, {
    "user-id": userId,
  });
}

export default {
  list,
  retrieve,
  create,
  invalidateBid,
  unsubscribe,
  subscribe,
  userBids,
  itemImages,
  itemBids,
  placeBid,
  itemWinningBid,
};
