const root =
  process.env.NODE_ENV === "production"
    ? "/api/v0.1.0"
    : "https://localhost:9000/api/v0.1.0";

const apiPublicURL =
  process.env.NODE_ENV === "production"
    ? process.env.PUBLIC_URL
    : "https://localhost:9000";

export default {
  apiPublicURL,
  apiRoot: root,
  itemsEndpoint: `${root}/items`,
  usersEndpoint: `${root}/accounts`,
  bidsEndpoint: `${root}/bids`,
  pageSize: 10,
};
