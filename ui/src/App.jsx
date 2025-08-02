import { useState } from 'react'
import reactLogo from './assets/react.svg'
import viteLogo from '/vite.svg'
import './App.css'
import OptionChainTable from './components/OptionChainData'
import StrikePriceChart from './components/charts'


function App() {
  const [count, setCount] = useState(0)

  return (
    <>

      <OptionChainTable />
    </>
  )
}

export default App
