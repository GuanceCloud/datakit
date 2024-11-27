import { createBrowserRouter, Navigate, useRoutes } from 'react-router-dom';

import Dashboard from '../pages/Dashboard/Dashboard';
import DkInfo from 'src/components/DkInfo/DkInfo';
import RunInfo from 'src/components/DkInfo/RunInfo/RunInfo';
import DatakitList from 'src/components/DatakitList/DkList';
import Pipeline from 'src/components/DkInfo/Pipeline/Pipeline';
import Config from 'src/components/DkInfo/Config/Config';
import Log from 'src/components/DkInfo/Log/Log';
import BlackList from 'src/components/DkInfo/BlackList/BlackList';
import Login from 'src/components/Login/Login';

function Routes() {
  const routes = useRoutes(
    [
      {
        path: '/dashboard',
        element: <Dashboard />,
        children: [
          {
            path: "",
            element: <DatakitList />
          },
          {
            path: '*',
            element: <DkInfo />,
            children: [
              {
                path: "runinfo",
                element: <RunInfo />
              },
              {
                path: "config",
                element: <Config />
              },
              {
                path: "pipeline",
                element: <Pipeline />
              },
              {
                path: "blacklist",
                element: <BlackList />
              },
              {
                path: "log",
                element: <Log />
              }
            ]
          },
        ]
      },
      {
        path: "login",
        element: <Login />
      },
      {
        path: "*",
        element: <Navigate to="/dashboard" />,
      }
    ]
  )

  return routes
}

const router = createBrowserRouter([
  {
    path: "*",
    element: <Routes />,
  },
])

export default router