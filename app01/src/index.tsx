import React, { useEffect, useState } from 'react';
import ReactDOM from 'react-dom';
import api from './api';
import { keycloak } from "./keycloak";




const App = () => {
  const [state, setState] = useState<any>({});
  const loadData = async () => {
    const resp = await api.getPageData();
    console.log(resp);
  }

  useEffect(() => {
    keycloak.init({onLoad: 'login-required'}).then(authenticated => {
      setState({ authenticated })
    })
  }, []);

  useEffect(() => {
    if (state.authenticated) {
      loadData();
    }
  }, [state.authenticated]);
  
  return <div>
    <div>App</div>

  </div>
}

const redirectUrl = window.location.origin;
ReactDOM.render(
  <React.StrictMode>

      <App />

  </React.StrictMode>,
  document.getElementById('root')
);
