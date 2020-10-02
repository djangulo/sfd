import React from "react";
import ReactDOM from "react-dom";
import { BrowserRouter as Router } from "react-router-dom";
import { createStore, applyMiddleware, compose } from "redux";
import App from "./App";
import * as serviceWorker from "./serviceWorker";

import thunk from "redux-thunk";
import { Provider } from "react-redux";

import { rootReducer } from "./store";

import { Loader, Dimmer } from "semantic-ui-react";

import "semantic-ui-css/semantic.min.css";
const composeEnhancers = window.__REDUX_DEVTOOLS_EXTENSION_COMPOSE__ || compose;

const configureStore = () => {
  const store = createStore(
    rootReducer,
    composeEnhancers(applyMiddleware(thunk))
  );

  return store;
};

const store = configureStore();

import("./i18n").then(() => {
  ReactDOM.render(
    <React.StrictMode>
      <Provider store={store}>
        <Router>
          <React.Suspense
            fallback={
              <Dimmer active>
                <Loader>Loading...</Loader>
              </Dimmer>
            }
          >
            <App />
          </React.Suspense>
        </Router>
      </Provider>
    </React.StrictMode>,
    document.getElementById("root")
  );
});

// If you want your app to work offline and load faster, you can change
// unregister() to register() below. Note this comes with some pitfalls.
// Learn more about service workers: https://bit.ly/CRA-PWA
serviceWorker.unregister();
