/** @jsx jsx */
// eslint-disable-next-line
import React from "react";
import { jsx } from "@emotion/core";
import { Header } from "semantic-ui-react";
import { useTranslation } from "react-i18next";
import ChangeLanguage from "./ChangeLanguage";
import HeaderMenu from "./HeaderMenu";

const SiteHeader = () => {
  const { t } = useTranslation("translation");
  return (
    <div
      css={{
        padding: "2rem",
        paddingBottom: "0.5rem",
        display: "flex",
        justifyContent: "space-between",
        backgroundColor: "cadetblue",
      }}
    >
      <Header as="h2">
        {t("translation:title.header")}
        <Header.Subheader>{t("translation:title.subheader")}</Header.Subheader>
      </Header>
      <HeaderMenu></HeaderMenu>
    </div>
  );
};

export default SiteHeader;
