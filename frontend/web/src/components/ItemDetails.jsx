/** @jsx jsx */
import { React } from "react";
import { jsx } from "@emotion/core";
import { connect } from "react-redux";
import { Card, Image, Tab, Header, List, Icon } from "semantic-ui-react";
import {
  itemIDFromSlug,
  itemsIsLoading,
  selectedItem,
  selectItem,
  getItemWinningBid,
} from "../store";
import { useEffect } from "react";
import LoaderWithText from "./common/LoaderWithText";
import { useParams } from "react-router-dom";
import { itemImagePath } from "../utils/index";
import { useTranslation } from "react-i18next";
import { getItemBids } from "../store/itemsDuck";

const ItemDetails = ({
  itemID = null,
  item = null,
  isLoading,
  selectItem,
  getItemWinningBid,
}) => {
  let { slug } = useParams();
  const { t, i18n } = useTranslation("item-details");
  useEffect(() => {
    if (item === null) {
      if (itemID === null) {
        selectItem(slug);
      } else {
        selectItem(itemID);
      }
    }
  }, [selectItem, item]);

  useEffect(() => {
    if (item && !item.blind) {
      getItemBids(item.id, { limit: 20 });
    }
  }, [getItemBids, item]);

  const parseDateTime = (lng, dt) =>
    dt.toLocaleDateString(lng) + " " + dt.toLocaleTimeString(lng);

  return isLoading ? (
    <LoaderWithText />
  ) : item ? (
    <Tab
      renderActiveOnly={true}
      panes={[
        {
          menuItem: t("item-details:Details"),
          render: () => (
            <Tab.Pane key="Details">
              <Header content={item.name} />
              <List>
                <List.Item>
                  <List.Header>{t("item-details:Description")}</List.Header>
                  <List.Description>{item.description}</List.Description>
                </List.Item>
                <List.Item>
                  <List.Header>{t("item-details:Closing date")}</List.Header>
                  <List.Description>
                    {parseDateTime(i18n.language, new Date(item.bid_deadline))}
                  </List.Description>
                </List.Item>
                <List.Item>
                  <List.Header>{t("item-details:Starting price")}</List.Header>
                  <List.Description>{item.starting_price}</List.Description>
                </List.Item>
                {item.max_price > 0 ? (
                  <List.Item>
                    <List.Header>{t("item-details:Max price")}</List.Header>
                    <List.Description>{item.max_price}</List.Description>
                  </List.Item>
                ) : null}
                <List.Item>
                  <List.Header>{t("item-details:Blind")}</List.Header>
                  <List.Description>
                    <Icon name={item.blind ? "check" : "close"} />
                  </List.Description>
                </List.Item>
                {!item.blind ? (
                  <List.Item>
                    <List.Header>{t("item-details:Winning bid")}</List.Header>
                    <List.Description>
                      {item.winning_bid ? item.winning_bid.amount : "0.00"}
                    </List.Description>
                  </List.Item>
                ) : null}
              </List>
            </Tab.Pane>
          ),
        },
        {
          menuItem: t("item-details:Bids"),
          render: () => (
            <Tab.Pane key="Bids">
              <Header
                content={t("item-details:Bids for", {
                  name: item.name,
                  default: `Bids for ${item.name}`,
                })}
              />
            </Tab.Pane>
          ),
        },
      ]}
    />
  ) : (
    <p>Not found</p>
  );
};

export default connect(
  (state) => ({
    isLoading: itemsIsLoading(state),
    item: selectedItem(state),
    idFromSlug: itemIDFromSlug(state),
  }),
  { selectItem, getItemWinningBid }
)(ItemDetails);
