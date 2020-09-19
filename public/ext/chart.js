function qsa(selector, parent) {
  if (!parent) { parent = document; }

  return Array.prototype.slice.call(parent.querySelectorAll(selector));
}

function element(name, content, attributes) {
  var e = document.createElement(name);

  if (content) { e.append(content); }
  for (var a in attributes) { e.setAttribute(a, attributes[a]); }

  return e;
}

function capitalise(string) {
  return string.charAt(0).toUpperCase() + string.slice(1);
}

// http://stackoverflow.com/questions/2901102/2901298#2901298
function thousands(x) {
  return x.toString().replace(/\B(?=(?:\d{3})+(?!\d))/g, ',');
}

function dataForAxis(data, yAxis) {
  return data.map(function(x) { return { t: x.t, y: x[yAxis] }; });
}

function createChart(data, yAxis) {
  return new Chart(document.querySelector('#chart'), {
    type: 'bar',
    data: {
      datasets: [{
        backgroundColor: '#edc240',
        data: dataForAxis(data, yAxis)
      }]
    },
    options: {
      events: ['click'],
      onClick: function(e, elements) {
        if (!elements.length) { return; }

        location.href = data[elements[0]._index].anchor;
      },
      legend: { display: false },
      scales: {
        xAxes: [{
          type: 'time',
          time: { unit: 'month' }
        }],
        yAxes: [{
          ticks: { callback: thousands }
        }]
      }
    }
  });
}

function setChartTitle(chart, data, yAxis) {
  var header = document.querySelector('#charts h2');
  var newY = (yAxis === 'books') ? 'pages' : 'books';
  var altY = element('span', newY + ' read', { class: 'click' });

  altY.addEventListener('click', function() {
    setChartTitle(chart, data, newY);
    chart.data.datasets[0].data = dataForAxis(data, newY);
    chart.update();
  });

  header.innerText = capitalise(yAxis) + ' read';
  header.append(' (show ', altY, ')');
}

function bookInfoItem(elem, item, func) {
  var definition = func ? func(elem) : elem.dataset[item];

  if (definition === '') { return []; }

  return [element('dt', capitalise(item)), element('dd', definition)];
};

function stars(elem) {
  var root = document.querySelector('h1 a').getAttribute('href');
  var rating = parseInt(elem.dataset.rating);

  if (rating === 0) { return ''; }

  return new Array(rating).fill('⭐️', 0, rating).join('');
}

function averageRating(elem) {
  return elem.dataset.average + ' (' + thousands(parseInt(elem.dataset.ratings)) + ')';
}

// Remove jQuery for this + tooltips
document.addEventListener('DOMContentLoaded', function () {
  var chartData = qsa('table.calendar').map(function(month) {
    var reviews = qsa('.reviews a', month);

    return {
      t: month.querySelector('.date').dataset.date,
      anchor: '#' + month.id,
      books: reviews.length,
      pages: reviews
        .map(function(link) { return +link.dataset.pages; })
        .reduce(function(a, x) { return a + x; }, 0)
    };
  });

  var yAxis = 'books';
  var chart = createChart(chartData, yAxis);

  setChartTitle(chart, chartData, yAxis);

  var userName = document.querySelector('#user-name').innerText.trim();

  tippy('td div.reviews a', {
    content: function(el) {
      var bookInfo = element('dl');
      var append = function(elements) { bookInfo.append.apply(bookInfo, elements); };

      ['title', 'author', 'published', 'pages'].forEach(function(field) {
        append(bookInfoItem(el, field));
      });

      append(bookInfoItem(el, userName + "’s rating", stars));
      append(bookInfoItem(el, 'overall rating', averageRating));
      bookInfo.append(element('div', '', { class: 'clear-left' }));

      return bookInfo;
    },
    allowHTML: true
  });
});
