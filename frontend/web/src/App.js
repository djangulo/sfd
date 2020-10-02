import React from "react";
import logo from "./logo.svg";
import "./App.css";
import HeaderMenu from "./components/HeaderMenu";
import Switcher from "./components/Switcher";

function App() {
  return (
    <div className="App">
      <HeaderMenu></HeaderMenu>
      <Switcher />
      {/* <header className="App-header">
        <img src={logo} className="App-logo" alt="logo" />
        <p>
          Edit <code>src/App.js</code> and save to reload.
        </p>
        <a
          className="App-link"
          href="https://reactjs.org"
          target="_blank"
          rel="noopener noreferrer"
        >
          Learn React
        </a>
      </header> */}
    </div>
  );
}

export default App;
