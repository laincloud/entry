import React from 'react';
import {
  render
} from 'react-dom';
import {
  Provider
} from 'react-redux';
import {
  createStore
} from 'redux';

import entryApp from './reducers'
import App from './containers/App'
import './index.css';
import registerServiceWorker from './registerServiceWorker';

const store = createStore(entryApp);

render(
  <Provider store={store}>
    <App />
  </Provider>,
  document.getElementById('root')
);

registerServiceWorker();
