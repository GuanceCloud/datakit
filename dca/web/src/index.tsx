import { createRoot } from 'react-dom/client';
import moment from 'moment';
import { Provider } from 'react-redux'
import { PersistGate } from 'redux-persist/integration/react';

import store, { persistor } from './store/index';
import './index.scss';
import App from './App';
import config from './config';

if (config.brandName === 'guance') {
  moment.locale('zh-cn')
}

const container = document.getElementById('root')
const root = createRoot(container!)

root.render(
  <Provider store={store}>
    <PersistGate persistor={persistor}>
      <App />
    </PersistGate>
  </Provider>
)
