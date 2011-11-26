var bookFields = ['date', 'title', 'author', 'pages', 'rating',
                  'published'];

// https://developer.mozilla.org/en/JavaScript/Reference/Global_Objects/Array/indexOf#Compatibility
if (!Array.prototype.indexOf) {
    Array.prototype.indexOf = function (searchElement /*, fromIndex */ ) {
        "use strict";
        if (this === void 0 || this === null) {
            throw new TypeError();
        }
        var t = Object(this);
        var len = t.length >>> 0;
        if (len === 0) {
            return -1;
        }
        var n = 0;
        if (arguments.length > 0) {
            n = Number(arguments[1]);
            if (n !== n) { // shortcut for verifying if it's NaN
                n = 0;
            } else if (n !== 0 && n !== Infinity && n !== -Infinity) {
                n = (n > 0 || -1) * Math.floor(Math.abs(n));
            }
        }
        if (n >= len) {
            return -1;
        }
        var k = n >= 0 ? n : Math.max(len - Math.abs(n), 0);
        for (; k < len; k++) {
            if (k in t && t[k] === searchElement) {
                return k;
            }
        }
        return -1;
    }
}

String.prototype.capitalise = function() {
    return this.charAt(0).toUpperCase() + this.slice(1);
}

function toDate(str) {
    dateParts = str.split('-');

    return new Date(Date.UTC(dateParts[0],
                             parseInt(dateParts[1], 10) - 1,
                             dateParts[2]));
}

function loadBookData() {
    return $('table.calendar td').map(function(i, cell) {
        cell = $(cell);

        if (cell.find('div').length == 0) { return null; }

        var date = toDate($(cell.find('.date')[0]).data('date'));

        var books = cell.find('.reviews a').map(function(j, link) {
            var data = { url: $(link).attr('href') };

            $.each(bookFields, function(k, field) {
                data[field] = $(link).data(field);
            });

            return data;
        });

        return [[date, $.makeArray(books)]];
    });
}

function labelFor(format) {
    return function(p) { return $.plot.formatDate(p[0], format); }
}

function addChartTitle(data, xAxis, yAxis) {
    var header = $('#charts h2');
    var click = { class: 'click' };
    var newX = (xAxis == 'month') ? 'day' : 'month';
    var newY = (yAxis == 'count') ? 'pages' : 'count';

    var chartTitle = yTitle(yAxis) + ' read by ' + xAxis;
    var altX = element('span', 'by ' + newX, click);
    var altY = element('span', yTitle(newY) + ' read', click);

    altX.click(function() { chart(data, newX, yAxis); });
    altY.click(function() { chart(data, xAxis, newY); });

    header.text(chartTitle.capitalise());

    header.
        append(' (show ').
        append(altY).
        append(' / ').
        append(altX).
        append(')');
}

function makeTicks(data, labeller) {
    var tickedOff = [];

    return $.map(data, function(point, i) {
        var tickLabel = labeller(point);

        if (tickedOff.indexOf(tickLabel) > -1) { return null; }

        tickedOff.push(tickLabel);

        return [[tickedOff.length - 1, tickLabel, point[0]]];
    });
}

function yTitle(y) { return (y == 'count') ? 'books' : y; }

function chartSwitches(xAxis, yAxis) {
    var ySize = function(book) {
        return (yAxis == 'count') ? 1 : parseInt('0' + book[yAxis], 10);
    }

    return {
        month: {
            dateFormat: '%y/%0m',
            ySize: ySize,
            xPoint: function(x, n) { return n; },
            xAxis: function(ts) {
                return { ticks: ts, tickLength: 0 };
            }
        },
        day: {
            dateFormat: '%y/%0m/%0d',
            ySize: ySize,
            xPoint: function(t, x) { return t[2]; },
            xAxis: function(x) {
                return { mode: 'time', timeformat: '%y/%0m' };
            }
        }
    }[xAxis];
}

function chart(data, xAxis, yAxis) {
    data = (typeof data == 'undefined') ? bookData : data;
    xAxis = (typeof xAxis == 'undefined') ? 'month' : xAxis;
    yAxis = (typeof yAxis == 'undefined') ? 'pages' : yAxis;

    var switches = chartSwitches(xAxis, yAxis);
    var label = labelFor(switches.dateFormat);
    var ticks = makeTicks(data, label);

    var chartData = $.map(ticks, function(tick, i) {
        var y = 0;

        var points = $.grep(data, function(point) {
            return (label(point) == tick[1]);
        });

        $.each(points, function(j, books) {
            $.each(books[1], function(k, book) {
                y += switches.ySize(book);
            });
        });

        return [[switches.xPoint(tick, i), y]];
    });

    addChartTitle(data, xAxis, yAxis);

    $.plot($('#chart'), [
        {
            data: chartData,
            bars: { show: true, align: 'center' }
        }
    ], {
        xaxis: switches.xAxis(ticks),
        yaxis: { tickFormatter: thousands, tickLength: 0 },
        grid: { clickable: true }
    });

    $('#chart').bind('plotclick', function(event, pos, item) {
        if (item) {
            monthID = ticks[item.dataIndex][1].replace(/\//g, '-');
            location.href = '#calendar-' + monthID;
        }
    });
};

var bookData = loadBookData();

$(function() { chart(); });
