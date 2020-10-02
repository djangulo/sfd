/** @jsx jsx */
import React, { useState } from "react";
import { useTranslation } from "react-i18next";

import { jsx, css } from "@emotion/core";

import { Header, Icon, List, Segment } from "semantic-ui-react";

const About = () => {
  const { t, i18n } = useTranslation("about");

  const founded = new Date("1992-08-12").toLocaleDateString(i18n.language, {
    year: "numeric",
    month: "long",
    day: "numeric",
  });

  const contactPhone = "8091234567";
  const displayPhone = (p) =>
    `(${p.slice(0, 3)})${p.slice(3, 6)}-${p.slice(6, 10)}`;

  return (
    <Segment id="about-segment">
      <Header as="h1">{t("about:About")}</Header>
      <p>
        {t(
          "The Dominican Philately Society (Sociedad Filatélica Dominicana) is a non-lucrative organization of volunteers",
          { dot: "." }
        )}
      </p>
      <List divided verticalAlign="middle">
        <List.Item>
          <List.Content floated="right">
            {t("Santo Domingo, DN Dominican Republic")}
          </List.Content>
          <List.Content floated="left">
            <Icon name="marker" />
            {t("about:Location")}
          </List.Content>
        </List.Item>
        <List.Item>
          <List.Content floated="right">{founded}</List.Content>
          <List.Content floated="left">
            <Icon name="calendar alternate" />
            {t("about:Founded")}
          </List.Content>
        </List.Item>
      </List>
      <Header as="h3">{t("Contact")}</Header>
      <List divided>
        <List.Item>
          <List.Content floated="right" as="a" href={`tel:${contactPhone}`}>
            {displayPhone(contactPhone)}
          </List.Content>
          <List.Content floated="left">
            <Icon name="phone" />
            {t("about:Phone")}
          </List.Content>
        </List.Item>
        <List.Item>
          <List.Content floated="right" as="a" href={`mailto:admin@sfd.com`}>
            admin@sfd.com
          </List.Content>
          <List.Content floated="left">
            <Icon name="mail" />
            {t("about:Email")}
          </List.Content>
        </List.Item>
      </List>
    </Segment>
  );
};

export default React.memo(About);
