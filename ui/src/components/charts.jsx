import React, { useState, useEffect, useMemo } from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { Card, Select, Typography } from 'antd';

const { Option } = Select;
const { Title } = Typography;

const OILineChart = ({ data = [] }) => {
  const [selectedExpiry, setSelectedExpiry] = useState('');
  const [selectedStrike, setSelectedStrike] = useState('');

  // Extract unique expiry dates and strike prices
  const { expiryDates, strikePrices, processedData } = useMemo(() => {
    if (!Array.isArray(data) || data.length === 0) {
      return { expiryDates: [], strikePrices: [], processedData: [] };
    }

    const expirySet = new Set();
    const strikeSet = new Set();
    const timestampMap = new Map();

    // Process all data points
    data.forEach(snapshot => {
      if (!snapshot.data || !Array.isArray(snapshot.data)) return;
      
      const timestamp = snapshot.timestamp;
      
      snapshot.data.forEach(item => {
        expirySet.add(item.expiryDate);
        strikeSet.add(item.strikePrice);
        
        const key = `${timestamp}-${item.strikePrice}-${item.expiryDate}`;
        
        if (!timestampMap.has(key)) {
          timestampMap.set(key, {
            timestamp,
            strikePrice: item.strikePrice,
            expiryDate: item.expiryDate,
            underlyingValue: snapshot.underlyingValue,
            ceOI: null,
            peOI: null
          });
        }
        
        const record = timestampMap.get(key);
        if (item.CE) {
          record.ceOI = item.CE.openInterest;
        }
        if (item.PE) {
          record.peOI = item.PE.openInterest;
        }
      });
    });

    return {
      expiryDates: Array.from(expirySet).sort(),
      strikePrices: Array.from(strikeSet).sort((a, b) => a - b),
      processedData: Array.from(timestampMap.values())
    };
  }, [data]);

  // Set default expiry and strike to first available
  useEffect(() => {
    if (expiryDates.length > 0 && !selectedExpiry) {
      setSelectedExpiry(expiryDates[0]);
    }
  }, [expiryDates, selectedExpiry]);

  useEffect(() => {
    if (strikePrices.length > 0 && !selectedStrike) {
      // Find the middle strike price as default
      const middleIndex = Math.floor(strikePrices.length / 2);
      setSelectedStrike(strikePrices[middleIndex]);
    }
  }, [strikePrices, selectedStrike]);

  // Filter and format data for the selected expiry and strike
  const chartData = useMemo(() => {
    if (!selectedExpiry || !selectedStrike) return [];

    const filteredData = processedData.filter(item => 
      item.expiryDate === selectedExpiry && item.strikePrice === selectedStrike
    );
    
    // Group by timestamp and format for market hours (9:15 AM to 3:30 PM)
    const timestampGroups = {};
    filteredData.forEach(item => {
      const date = new Date(item.timestamp);
      const hours = date.getHours();
      const minutes = date.getMinutes();
      
      // Filter for market hours (9:15 AM to 3:30 PM)
      if ((hours === 9 && minutes >= 15) || (hours >= 10 && hours < 15) || (hours === 15 && minutes <= 30)) {
        // Round to nearest 3-minute interval
        const roundedMinutes = Math.floor(minutes / 3) * 3;
        const roundedTime = new Date(date);
        roundedTime.setMinutes(roundedMinutes, 0, 0);
        
        const timeKey = roundedTime.getTime();
        const formattedTime = roundedTime.toLocaleTimeString('en-IN', { 
          hour: '2-digit', 
          minute: '2-digit',
          hour12: false 
        });

        if (!timestampGroups[timeKey]) {
          timestampGroups[timeKey] = {
            timestamp: timeKey,
            formattedTime,
            underlyingValue: item.underlyingValue,
            callOI: null,
            putOI: null
          };
        }
        
        // Use the latest values for each time interval
        if (item.ceOI !== null) {
          timestampGroups[timeKey].callOI = item.ceOI;
        }
        if (item.peOI !== null) {
          timestampGroups[timeKey].putOI = item.peOI;
        }
      }
    });

    // Generate time intervals from 9:15 AM to 3:30 PM with 3-minute gaps
    const marketStart = new Date();
    marketStart.setHours(9, 15, 0, 0);
    const marketEnd = new Date();
    marketEnd.setHours(15, 30, 0, 0);
    
    const timeIntervals = [];
    const current = new Date(marketStart);
    
    while (current <= marketEnd) {
      const timeKey = current.getTime();
      const formattedTime = current.toLocaleTimeString('en-IN', { 
        hour: '2-digit', 
        minute: '2-digit',
        hour12: false 
      });
      
      if (timestampGroups[timeKey]) {
        timeIntervals.push(timestampGroups[timeKey]);
      } else {
        timeIntervals.push({
          timestamp: timeKey,
          formattedTime,
          underlyingValue: null,
          callOI: null,
          putOI: null
        });
      }
      
      current.setMinutes(current.getMinutes() + 3);
    }

    return timeIntervals;
  }, [processedData, selectedExpiry, selectedStrike]);

  // Custom tick formatter for X-axis to show only specific time intervals
  const formatXAxisTick = (tickItem) => {
    const time = tickItem;
    const date = new Date();
    const [hours, minutes] = time.split(':');
    date.setHours(parseInt(hours), parseInt(minutes));
    
    // Show ticks every 30 minutes
    if (parseInt(minutes) % 30 === 0 || (parseInt(hours) === 9 && parseInt(minutes) === 15)) {
      return time;
    }
    return '';
  };

  if (!data || data.length === 0) {
    return (
      <Card className="w-full">
        <div className="text-center py-8 text-gray-500">
          No data available for chart
        </div>
      </Card>
    );
  }

  return (
    <Card className="w-full shadow-lg rounded-2xl">
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center mb-4">
        <Title level={4} className="!mb-0">Open Interest Movement</Title>
        <div className="flex flex-col sm:flex-row gap-4 mt-2 sm:mt-0">
          <div className="flex items-center gap-2">
            <span className="font-semibold text-gray-600">Expiry:</span>
            <Select
              value={selectedExpiry}
              onChange={setSelectedExpiry}
              className="!w-48"
              size="small"
            >
              {expiryDates.map(expiry => (
                <Option key={expiry} value={expiry}>{expiry}</Option>
              ))}
            </Select>
          </div>
          <div className="flex items-center gap-2">
            <span className="font-semibold text-gray-600">Strike:</span>
            <Select
              value={selectedStrike}
              onChange={setSelectedStrike}
              className="!w-32"
              size="small"
            >
              {strikePrices.map(strike => (
                <Option key={strike} value={strike}>{strike}</Option>
              ))}
            </Select>
          </div>
        </div>
      </div>

      {chartData.length > 0 ? (
        <div className="w-full" style={{ height: '500px' }}>
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={chartData} margin={{ top: 5, right: 30, left: 20, bottom: 60 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
              <XAxis 
                dataKey="formattedTime"
                stroke="#666"
                fontSize={11}
                angle={-45}
                textAnchor="end"
                height={80}
                interval={0}
                tickFormatter={formatXAxisTick}
              />
              <YAxis 
                stroke="#666"
                fontSize={12}
                tickFormatter={(value) => value.toLocaleString()}
              />
              <Tooltip 
                formatter={(value, name) => [
                  value ? value.toLocaleString() : 'No data',
                  name
                ]}
                labelFormatter={(label) => `Time: ${label}`}
                contentStyle={{
                  backgroundColor: '#fff',
                  border: '1px solid #ccc',
                  borderRadius: '8px',
                  fontSize: '12px'
                }}
              />
              <Legend 
                wrapperStyle={{ fontSize: '14px', paddingTop: '20px' }}
                iconType="line"
              />
              
              {/* Call Option OI Line - Green */}
              <Line
                dataKey="callOI"
                stroke="#22c55e"
                strokeWidth={3}
                dot={false}
                name={`Call OI (${selectedStrike})`}
                connectNulls={false}
              />
              
              {/* Put Option OI Line - Red */}
              <Line
                dataKey="putOI"
                stroke="#ef4444"
                strokeWidth={3}
                dot={false}
                name={`Put OI (${selectedStrike})`}
                connectNulls={false}
              />
            </LineChart>
          </ResponsiveContainer>
        </div>
      ) : (
        <div className="text-center py-8 text-gray-500">
          No data available for selected expiry date and strike price
        </div>
      )}
      
      <div className="mt-4 text-sm text-gray-600 bg-gray-50 p-3 rounded-lg">
        <div className="flex flex-wrap gap-6">
          <div className="flex items-center gap-2">
            <div className="w-4 h-0.5 bg-green-500"></div>
            <span>Call Options Open Interest</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-4 h-0.5 bg-red-500"></div>
            <span>Put Options Open Interest</span>
          </div>
          <span className="text-xs text-gray-500">
            Market Hours: 9:15 AM - 3:30 PM | 3-minute intervals
          </span>
        </div>
      </div>
    </Card>
  );
};

export default OILineChart;