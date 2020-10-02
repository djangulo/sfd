import React from "react";
import { connect } from "react-redux";
import { useHistory } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { Dropdown } from "semantic-ui-react";
import { i18nGetKey, setRouteKey } from "../store";

const ChangeLanguage = ({ routeKey, setRouteKey }) => {
  const history = useHistory();
  const { t, i18n } = useTranslation(["router", "menu"]);
  const changeLanguage = (lng) => {
    i18n.changeLanguage(lng);
    history.push(t(`router:${routeKey}`));
    console.log(routeKey);
  };

  const getLanguageOptions = () =>
    [
      { name: "English", key: "en-US", flag: "us" },
      { name: "Español", key: "es-DO", flag: "do" },
    ].map((l) => ({
      key: l.key,
      value: l.key,
      text: l.name,
      flag: l.flag,
    }));

  return (
    <Dropdown
      compact
      button
      className="icon"
      floating
      labeled
      selection
      icon="world"
      text={t("menu:Language")}
      onChange={(_, data) => {
        console.log(data);
        changeLanguage(data.value);
      }}
      options={getLanguageOptions()}
    />
  );
};

export default connect(
  (state) => ({
    routeKey: i18nGetKey(state),
  }),
  { setRouteKey }
)(ChangeLanguage);
