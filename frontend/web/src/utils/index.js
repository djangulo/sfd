import config from "../services/config";

export const itemImagePath = (path) => `${config.apiPublicURL}${path}`;
