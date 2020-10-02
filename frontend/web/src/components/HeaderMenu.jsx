/** @jsx jsx */
import React from "react";
import { connect } from "react-redux";
import { Menu, Input, Button } from "semantic-ui-react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router-dom";
import ChangeLanguage from "./ChangeLanguage";
import { setRouteKey } from "../store";

import { jsx, css } from "@emotion/core";

import { MediaQueries } from "./appMedia";
import { useState } from "react";

const menuItems = ({ t, setRouteKey }) => {
  return (
    <React.Fragment>
      <Menu.Item
        as={Link}
        to={t("router:items.home.path")}
        onClick={() => setRouteKey("items.home.path")}
      >
        {t("router:items.home.title")}
      </Menu.Item>
      <Menu.Item
        as={Link}
        to={t("router:about.path")}
        onClick={() => setRouteKey("about.path")}
      >
        {t("router:about.title")}
      </Menu.Item>
      <Menu.Item>
        <Input icon="search" placeholder={`${t("menu:Search")}...`} />
      </Menu.Item>
      <Menu.Item
        as={Link}
        to={t("router:login.path")}
        onClick={() => setRouteKey("login.path")}
      >
        {t("router:login.title")}
      </Menu.Item>
      <Menu.Item
        as={Link}
        to={t("router:signup.path")}
        onClick={() => setRouteKey("signup.path")}
      >
        {t("router:signup.title")}
      </Menu.Item>
      <Menu.Item>
        <ChangeLanguage />
      </Menu.Item>
    </React.Fragment>
  );
};

const display = (showMenu) =>
  MediaQueries({
    display: [
      showMenu ? "flex" : "none",
      showMenu ? "flex" : "none",
      "flex",
      "flex",
    ],
    visibility: [
      showMenu ? "visible" : "hidden",
      showMenu ? "visible" : "hidden",
      "visible",
      "visible",
    ],
  });

const HeaderMenu = ({ setRouteKey }) => {
  const [showMenu, setShowMenu] = useState(false);
  const { t } = useTranslation(["menu", "router"]);
  return (
    <Menu
      secondary
      fluid
      css={MediaQueries({
        flexDirection: [
          showMenu ? "column" : "row-reverse",
          showMenu ? "column" : "row-reverse",
          "row",
          "row",
        ],
        height: !showMenu ? "3em" : "inherit",
        backgroundColor: "#a3a3a3",
      })}
    >
      <Menu.Item as={Link} to="/">
        SFD
      </Menu.Item>
      <Menu.Item
        position="right"
        css={MediaQueries({
          display: ["flex", "flex", "none", "none"],
          visibility: ["visible", "visible", "hidden", "hidden"],
        })}
      >
        <Button
          size="big"
          icon={showMenu ? "close" : "bars"}
          floated="right"
          onClick={() => setShowMenu(!showMenu)}
        />
      </Menu.Item>
      {/* <Menu.Menu
        vertical={showMenu}
        css={MediaQueries({
          display: [
            showMenu ? "flex" : "none",
            showMenu ? "flex" : "none",
            "flex",
            "flex",
          ],
          visibility: [
            showMenu ? "visible" : "hidden",
            showMenu ? "visible" : "hidden",
            "visible",
            "visible",
          ],
          width: [showMenu ? "100%" : 0, showMenu ? "100%" : 0, "100%", "100%"],
        })}
      > */}
      <Menu.Item
        as={Link}
        to={t("router:items.home.path")}
        onClick={() => {
          setRouteKey("items.home.path");
          setShowMenu(!showMenu);
        }}
        css={display(showMenu)}
      >
        {t("router:items.home.title")}
      </Menu.Item>
      <Menu.Item
        as={Link}
        to={t("router:about.path")}
        onClick={() => {
          setRouteKey("about.path");
          setShowMenu(!showMenu);
        }}
        css={display(showMenu)}
      >
        {t("router:about.title")}
      </Menu.Item>
      <Menu.Menu
        // position="right"
        css={MediaQueries({
          flexDirection: ["column", "column", "row", "row"],
          alignSelf: "center",
        })}
      >
        <Menu.Item
          css={
            (display(showMenu),
            MediaQueries({
              order: [1, 1, "inherit", "inherit"],
            }))
          }
        >
          <Input
            icon="search"
            placeholder={`${t("menu:Search")}...`}
            css={display(showMenu)}
          />
        </Menu.Item>
        <Menu.Item
          as={Link}
          to={t("router:login.path")}
          onClick={() => {
            setRouteKey("login.path");
            setShowMenu(!showMenu);
          }}
          css={display(showMenu)}
        >
          {t("router:login.title")}
        </Menu.Item>
        <Menu.Item
          as={Link}
          to={t("router:signup.path")}
          onClick={() => {
            setRouteKey("signup.path");
            setShowMenu(!showMenu);
          }}
          css={display(showMenu)}
        >
          {t("router:signup.title")}
        </Menu.Item>
        <Menu.Item css={display(showMenu)}>
          <ChangeLanguage />
        </Menu.Item>
      </Menu.Menu>
      {/* </Menu.Menu> */}
    </Menu>
  );
};
export default connect(null, { setRouteKey })(HeaderMenu);
