import React from "react";
import { useTranslation } from "react-i18next";
import { Switch, Route, Redirect } from "react-router-dom";
import loadable from "@loadable/component";
import { Container } from "semantic-ui-react";

const LoadableAbout = loadable(() => import("./About.jsx"));
const LoadableItemList = loadable(() => import("./ItemList.jsx"));
const LoadableItemDetail = loadable(() => import("./ItemDetails.jsx"));

const Switcher = () => {
  const { t } = useTranslation();
  const languages = ["en-US", "es-DO"];
  const routes = [
    { key: "items.detail.path", component: LoadableItemDetail },
    { key: "items.home.path", component: LoadableItemList },
    // {key: "items.new.path", component: },
    { key: "about.path", component: LoadableAbout },
    // {key: "login.path", component: },
    // {key: "account.path", component: },
  ];
  return (
    <Container>
      <Switch>
        <Route exact path="/">
          {<Redirect to={t("router:items.home.path")} />}
        </Route>
        {routes.map((route) => (
          <Route path={t(route.key)} key={route.key}>
            {<route.component />}
          </Route>
        ))}
      </Switch>
    </Container>
  );
};

export default Switcher;
