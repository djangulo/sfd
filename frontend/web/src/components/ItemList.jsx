import React from "react";
import { connect } from "react-redux";

import { NavLink } from "react-router-dom";

import { useTranslation } from "react-i18next";
import { Item, Loader } from "semantic-ui-react";
import {
  selectedItem,
  setSelectedItem,
  itemsIsLoading,
  fetchItems,
  getItems,
  itemsLastID,
  itemsLastCreatedAt,
} from "../store";
import { itemImagePath } from "../utils/index";

const ItemList = ({
  isLoading,
  setSelectedItem,
  fetchItems,
  allItems,
  lastID,
  lastCreated,
}) => {
  const { t, i18n } = useTranslation(["router", "items"]);
  React.useEffect(() => {
    var opts = {};
    if (lastID !== null) {
      opts.lastID = lastID;
    }
    if (lastCreated !== null) {
      opts.lastCreatedAt = lastCreated;
    }
    fetchItems(opts);
  }, [fetchItems]);

  const parseDateTime = (lng, dt) =>
    dt.toLocaleDateString(lng) + " " + dt.toLocaleTimeString(lng);
  return (
    <>
      <Item.Group id="item-list" divided relaxed>
        {allItems.map((item) => (
          <Item
            key={item.id}
            className="item"
            as={NavLink}
            to={t("router:items.detail.url", { slug: item.slug })}
            onClick={() => setSelectedItem(item)}
          >
            <Item.Image
              size="tiny"
              src={itemImagePath(item.cover_image.path)}
              alt={item.cover_image.alt_text}
            />
            <Item.Content>
              <Item.Header>{item.name}</Item.Header>
              <Item.Meta>
                <span className="cinema">
                  {`${t("items:Closes at")} ${parseDateTime(
                    i18n.language,
                    new Date(item.bid_deadline)
                  )}`}
                </span>
              </Item.Meta>
              <Item.Description>{item.description}</Item.Description>
            </Item.Content>
          </Item>
        ))}
      </Item.Group>
      {isLoading ? <Loader active inline="centered" /> : null}
    </>
  );
};

export default connect(
  (state) => ({
    isLoading: itemsIsLoading(state),
    selectedItem: selectedItem(state),
    allItems: getItems(state),
    lastID: itemsLastID(state),
    lastCreated: itemsLastCreatedAt(state),
  }),
  { fetchItems, setSelectedItem }
)(ItemList);
