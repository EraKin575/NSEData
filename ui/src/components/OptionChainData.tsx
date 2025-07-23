import { useEffect, useState } from "react";
import axios from "axios";
import { Table, Card, Spin, Alert, InputNumber, Select, Button } from "antd";

const { Option } = Select;

// Define columns with CE (call) and PE (put) side by side for each attribute
const columns = [
  {
    title: "Strike Price",
    dataIndex: "strikePrice",
    key: "strikePrice",
    fixed: "left",
    render: (v) => v?.toLocaleString(),
    sorter: (a, b) => a.strikePrice - b.strikePrice,
    width: 100,
  },
  {
    title: "Expiry Date",
    dataIndex: "expiryDate",
    key: "expiryDate",
    sorter: (a, b) => a.expiryDate.localeCompare(b.expiryDate),
    width: 120,
  },
  {
    title: "CE OI",
    dataIndex: "ceOpenInterest",
    key: "ceOpenInterest",
    render: (v) => v !== undefined ? v.toLocaleString() : "-",
    width: 80,
  },
  {
    title: "CE Chg OI",
    dataIndex: "ceChangeInOI",
    key: "ceChangeInOI",
    render: (v) => v !== undefined ? v.toLocaleString() : "-",
    width: 90,
  },
  {
    title: "CE % Chg OI",
    dataIndex: "cePChangeInOI",
    key: "cePChangeInOI",
    render: (v) => v !== undefined ? v.toFixed(2) + "%" : "-",
    width: 100,
  },
  {
    title: "CE Vol",
    dataIndex: "ceVolume",
    key: "ceVolume",
    render: (v) => v !== undefined ? v.toLocaleString() : "-",
    width: 90,
  },
  {
    title: "CE IV",
    dataIndex: "ceIV",
    key: "ceIV",
    render: (v) => v !== undefined ? v.toFixed(2) : "-",
    width: 70,
  },
  {
    title: "CE LTP",
    dataIndex: "ceLTP",
    key: "ceLTP",
    render: (v) => v !== undefined ? v.toFixed(2) : "-",
    width: 80,
  },
  {
    title: "PE LTP",
    dataIndex: "peLTP",
    key: "peLTP",
    render: (v) => v !== undefined ? v.toFixed(2) : "-",
    width: 80,
  },
  {
    title: "PE IV",
    dataIndex: "peIV",
    key: "peIV",
    render: (v) => v !== undefined ? v.toFixed(2) : "-",
    width: 70,
  },
  {
    title: "PE Vol",
    dataIndex: "peVolume",
    key: "peVolume",
    render: (v) => v !== undefined ? v.toLocaleString() : "-",
    width: 90,
  },
  {
    title: "PE % Chg OI",
    dataIndex: "pePChangeInOI",
    key: "pePChangeInOI",
    render: (v) => v !== undefined ? v.toFixed(2) + "%" : "-",
    width: 100,
  },
  {
    title: "PE Chg OI",
    dataIndex: "peChangeInOI",
    key: "peChangeInOI",
    render: (v) => v !== undefined ? v.toLocaleString() : "-",
    width: 90,
  },
  {
    title: "PE OI",
    dataIndex: "peOpenInterest",
    key: "peOpenInterest",
    render: (v) => v !== undefined ? v.toLocaleString() : "-",
    width: 80,
  },
];

function OptionChainTable() {
  const [loading, setLoading] = useState(true);
  const [rawRecords, setRawRecords] = useState([]);
  const [records, setRecords] = useState([]);
  const [meta, setMeta] = useState({ timestamp: "", underlyingValue: 0 });
  const [error, setError] = useState(null);

  const [minStrike, setMinStrike] = useState(null);
  const [maxStrike, setMaxStrike] = useState(null);
  const [expiryDates, setExpiryDates] = useState([]);
  const [selectedExpiry, setSelectedExpiry] = useState(undefined);

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 180000);
    return () => clearInterval(interval);
    // eslint-disable-next-line
  }, []);

  const fetchData = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await axios.post("http://localhost:4300/api/data", {}); // adjust endpoint if needed
      // Map by strikePrice + expiryDate
      const map = {};
      const expirySet = new Set();
      res.data.data.forEach((item) => {
        const key = item.strikePrice + "-" + item.expiryDate;
        expirySet.add(item.expiryDate);
        if (!map[key]) {
          map[key] = {
            key,
            strikePrice: item.strikePrice,
            expiryDate: item.expiryDate,
          };
        }
        if (item.CE) {
          map[key].ceOpenInterest = item.CE.openInterest;
          map[key].ceChangeInOI = item.CE.changeinOpenInterest;
          map[key].cePChangeInOI = item.CE.pchangeinOpenInterest;
          map[key].ceVolume = item.CE.totalTradedVolume;
          map[key].ceIV = item.CE.impliedVolatility;
          map[key].ceLTP = item.CE.lastPrice;
        }
        if (item.PE) {
          map[key].peOpenInterest = item.PE.openInterest;
          map[key].peChangeInOI = item.PE.changeinOpenInterest;
          map[key].pePChangeInOI = item.PE.pchangeinOpenInterest;
          map[key].peVolume = item.PE.totalTradedVolume;
          map[key].peIV = item.PE.impliedVolatility;
          map[key].peLTP = item.PE.lastPrice;
        }
      });
      const allRecords = Object.values(map);
      setRawRecords(prev => [...prev, ...allRecords]);
      setExpiryDates([...expirySet].sort());
      setMeta({
        timestamp: res.data.timestamp,
        underlyingValue: res.data.underlyingValue,
      });
      setLoading(false);
      // Optionally auto-select first expiry
      if (!selectedExpiry && expirySet.size > 0) setSelectedExpiry([...expirySet][0]);
    } catch (err) {
      setError("Failed to fetch data from backend.");
      setLoading(false);
    }
  };

  // Filtering logic
  useEffect(() => {
    let filtered = rawRecords;
    if (selectedExpiry) {
      filtered = filtered.filter((rec) => rec.expiryDate === selectedExpiry);
    }
    if (minStrike !== null) {
      filtered = filtered.filter((rec) => rec.strikePrice >= minStrike);
    }
    if (maxStrike !== null) {
      filtered = filtered.filter((rec) => rec.strikePrice <= maxStrike);
    }
    setRecords(filtered);
  }, [rawRecords, minStrike, maxStrike, selectedExpiry]);

  return (
    <div className="w-full max-w-7xl mx-auto p-4">
      <Card
        className="mb-4 shadow-lg rounded-2xl"
        title={
          <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center">
            <span className="text-lg font-bold">NIFTY Option Chain</span>
            <span className="text-sm text-gray-500 mt-2 sm:mt-0">
              Last Updated: {meta.timestamp ? new Date(meta.timestamp).toLocaleString() : "--"}
            </span>
          </div>
        }
        extra={
          <span className="text-blue-700 font-semibold text-xl">
            Underlying Value: {meta.underlyingValue || "--"}
          </span>
        }
      >
        {/* Filters */}
        <div className="flex flex-col md:flex-row items-center gap-3 mb-4">
          <div className="flex items-center gap-2">
            <span className="font-semibold text-gray-600">Strike Price:</span>
            <InputNumber
              min={0}
              placeholder="Min"
              size="small"
              value={minStrike}
              onChange={setMinStrike}
              className="!w-24"
            />
            <span>-</span>
            <InputNumber
              min={0}
              placeholder="Max"
              size="small"
              value={maxStrike}
              onChange={setMaxStrike}
              className="!w-24"
            />
          </div>
          <div className="flex items-center gap-2">
            <span className="font-semibold text-gray-600">Expiry Date:</span>
            <Select
              showSearch
              size="small"
              placeholder="Select expiry"
              value={selectedExpiry}
              onChange={setSelectedExpiry}
              className="!w-48"
              optionFilterProp="children"
              allowClear={false}
            >
              {expiryDates.map((exp) => (
                <Option key={exp} value={exp}>
                  {exp}
                </Option>
              ))}
            </Select>
          </div>
          <Button
            size="small"
            onClick={() => {
              setMinStrike(null);
              setMaxStrike(null);
              if (expiryDates.length) setSelectedExpiry(expiryDates[0]);
            }}
          >
            Reset
          </Button>
        </div>

        {error && <Alert message={error} type="error" className="mb-4" />}
        {loading ? (
          <div className="flex justify-center items-center min-h-[200px]">
            <Spin size="large" />
          </div>
        ) : (
          <Table
            columns={columns}
            dataSource={records}
            size="small"
            pagination={{ pageSize: 25 }}
            scroll={{ x: "max-content" }}
            bordered
            rowClassName="text-xs"
            className="rounded-2xl"
            sticky
          />
        )}
      </Card>
    </div>
  );
}

export default OptionChainTable;
