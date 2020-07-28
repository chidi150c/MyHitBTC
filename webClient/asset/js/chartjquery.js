

                    var detailChart;
                    var num;
                         $(document).ready(function(){  
                             // Apply the theme
                    Highcharts.setOptions(Highcharts.theme);
                            // z = makeChartBody(chartData) 
                            requestData1()
                            requestData()
                        }); 

                        function requestData1() {
                            $.getJSON('/graphpoint', function (data) {                                                 
                                 num = Number(data[data.length - 1][1]).toLocaleString('en');
                              $('#BTCNGN').text(num);
                              // call it again after one second
                                setTimeout(requestData, 60 * 1000);
                          });
                      }

                        function requestData() {
                            $.getJSON('/graphpoint', function (data) {
                                  // make the container smaller and add a second container for the master chart
                                  var $container = $('#container')
                                  .css('position', 'relative');
                          
                              $('<div id="detail-container">')
                                  .appendTo($container);
                          
                              $('<div id="master-container">')
                                  .css({
                                      position: 'absolute',
                                      top: 300,
                                      height: 100,
                                      width: '100%'
                                  })
                                      .appendTo($container);
                          
                              // create master and in its callback, create the detail chart
                              createMaster(data);
                            
                             
                                 num = Number(data[data.length - 1][1]).toLocaleString('en');
                            
                              $('#BTCNGN').text(num)
                              // call it again after one second
                                  setTimeout(requestData, 2 * 60 * 1000);
                          });
                      }
                            
                                // create the detail chart
                                function createDetail(masterChart, data) {
                            
                                    // prepare the detail chart
                                    var detailData = [],
                                        detailStart = data[0][0];
                            
                                    $.each(masterChart.series[0].data, function () {
                                        if (this.x >= detailStart) {
                                            detailData.push(this.y);
                                        }
                                    });
                            
                                    // create a detail chart referenced by a global variable
                                    detailChart = Highcharts.chart('detail-container', {
                                        chart: {
                                            borderWidth: 1,
                                            plotBorderWidth: 1,
                                            marginBottom: 120,
                                            reflow: false,
                                            marginLeft: 90,
                                            marginRight: 10,
                                            style: {
                                                position: 'absolute'
                                            }
                                        },
                                        credits: {
                                            enabled: false
                                        },
                                        title: {
                                            text: 'NAIRA to BITCOIN Exchange Rate'
                                        },
                                        subtitle: {
                                            text: 'Select an area by dragging across the lower chart'
                                        },
                                        xAxis: {
                                            type: 'datetime'
                                        },
                                        yAxis: {
                                            title: {
                                                text: 'Price(₦)'
                                            },
                                            maxZoom: 0.1
                                        },
                                        tooltip: {
                                            formatter: function () {
                                                var point = this.points[0];
                                                return '<b>' + point.series.name + '</b><br/>' + Highcharts.dateFormat('%A %B %e %Y', this.x) + ':<br/>' +
                                                    '1 BTC = ' + '₦' + Highcharts.numberFormat(point.y, 2);
                                            },
                                            shared: true
                                        },
                                        legend: {
                                            enabled: false
                                        },
                                        plotOptions: {
                                            series: {
                                                marker: {
                                                    enabled: false,
                                                    states: {
                                                        hover: {
                                                            enabled: true,
                                                            radius: 3
                                                        }
                                                    }
                                                }
                                            }
                                        },
                                        series: [{
                                            name: 'NGN to BTC',
                                            pointStart: detailStart,
                                            //pointInterval: 24 * 3600 * 1000,
                                            data: detailData
                                        }],
                            
                                        exporting: {
                                            enabled: false
                                        }
                            
                                    }); // return chart
                                }
                            
                                // create the master chart
                                function createMaster(data) {                                
                                    Highcharts.chart('master-container', {
                                        chart: {
                                            reflow: false,
                                            borderWidth: 0,
                                            backgroundColor: null,
                                            marginLeft: 50,
                                            marginRight: 20,
                                            zoomType: 'x',
                                            events: {
                            
                                                // listen to the selection event on the master chart to update the
                                                // extremes of the detail chart
                                                selection: function (event) {
                                                    var extremesObject = event.xAxis[0],
                                                        min = extremesObject.min,
                                                        max = extremesObject.max,
                                                        detailData = [],
                                                        xAxis = this.xAxis[0];
                            
                                                    // reverse engineer the last part of the data
                                                    $.each(this.series[0].data, function () {
                                                        if (this.x > min && this.x < max) {
                                                            detailData.push([this.x, this.y]);
                                                        }
                                                    });
                            
                                                    // move the plot bands to reflect the new detail span
                                                    xAxis.removePlotBand('mask-before');
                                                    xAxis.addPlotBand({
                                                        id: 'mask-before',
                                                        from: data[0][0],
                                                        to: min,
                                                        color: 'rgba(0, 0, 0, 0.2)'
                                                    });
                            
                                                    xAxis.removePlotBand('mask-after');
                                                    xAxis.addPlotBand({
                                                        id: 'mask-after',
                                                        from: max,
                                                        to: data[data.length - 1][0],
                                                        color: 'rgba(0, 0, 0, 0.2)'
                                                    });
                            
                            
                                                    detailChart.series[0].setData(detailData);
                            
                                                    return false;
                                                }
                                            }
                                        },
                                        title: {
                                            text: null
                                        },
                                        xAxis: {
                                            type: 'datetime',
                                            showLastTickLabel: true,
                                            maxZoom: 2 * 24 * 3600000, // fourteen days 14 * 
                                            plotBands: [{
                                                id: 'mask-before',
                                                from: data[0][0],
                                                to: data[data.length - 1][0],
                                                color: 'rgba(0, 0, 0, 0.2)'
                                            }],
                                            title: {
                                                text: null
                                            }
                                        },
                                        yAxis: {
                                            gridLineWidth: 0,
                                            labels: {
                                                enabled: false
                                            },
                                            title: {
                                                text: null
                                            },
                                            min: 0.6,
                                            showFirstLabel: false
                                        },
                                        tooltip: {
                                            formatter: function () {
                                                return false;
                                            }
                                        },
                                        legend: {
                                            enabled: false
                                        },
                                        credits: {
                                            enabled: false
                                        },
                                        plotOptions: {
                                            series: {
                                                fillColor: {
                                                    linearGradient: [0, 0, 0, 70],
                                                    stops: [
                                                        [0, Highcharts.getOptions().colors[0]],
                                                        [1, 'rgba(255,255,255,0)']
                                                    ]
                                                },
                                                lineWidth: 1,
                                                marker: {
                                                    enabled: false
                                                },
                                                shadow: false,
                                                states: {
                                                    hover: {
                                                        lineWidth: 1
                                                    }
                                                },
                                                enableMouseTracking: false
                                            }
                                        },
                            
                                        series: [{
                                            type: 'area',
                                            name: 'NGN to BTC',
                                            //pointInterval: 24 * 3600 * 1000,
                                            pointStart: data[0][0],
                                            data: data
                                        }],
                            
                                        exporting: {
                                            enabled: false
                                        }
                            
                                    }, function (masterChart) {
                                        createDetail(masterChart, data);
                                    }); // return chart instance
                                }

                                