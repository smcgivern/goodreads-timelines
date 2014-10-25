require 'bundler/setup'
require './lib/goodreads'
require 'sinatra'
require 'haml'
require 'kramdown'

module Kramdown
  include Haml::Filters::Base

  def render(text)
    ::Kramdown::Document.new(text).to_html
  end
end

set(:views, Proc.new {File.join(root, 'view')})
set(:haml, :format => :html5)

ROOT = (File.exist?('.root') ?
        File.open('.root').read.strip : nil)

helpers do
  # Call a partial HAML template.
  def partial(template, variables={})
    haml(template, {:layout => false}, variables)
  end

  # Prepend ROOT (from .root, if it exists) to absolute paths.
  def r(s)
    (s =~ /^\// && !(s =~ /^\/#{ROOT}/)) ? "/#{ROOT}#{s}" : s
  end

  # Take the inner_text and remove surrounding whitespace from a
  # Nokogiri element.
  def t(e); e.inner_text.strip; end

  # Formats a number using a comma as the thousands seperator.
  def thousands(n)
    n.to_s.gsub(/(\d)(?=(\d\d\d)+(?!\d))/, "\\1,")
  end

  # Gets the important book information from a review and turns it
  # into a hash.
  def book_info(review)
    info = {
      :date => 'read_at',
      :title => 'title',
      :author => 'author name',
      :thumbnail => 'small_image_url',
      :pages => 'num_pages',
      :rating => 'rating',
      :average => 'average_rating',
      :ratings => 'ratings_count',
      :published => 'published',
    }

    Hash[info.map {|k, v| [k, t(review.at(v))]}]
  end
end

get('/ext/style.css') {scss(:style)}

get('/') do
  @page_title = 'Goodreads timelines'
  @scripts = ['/ext/jquery-1.7.min.js', '/ext/index.js']

  haml(:index)
end

post('/go-to-timeline/?') do
  user_id = Goodreads.find_user_id(params['goodreads-uri'])

  redirect(r("/:#{user_id}/"))
end

get('/::user_id/?') do
  @user_id = params['user_id'].match(/[0-9]+/).to_s
  @user_info = Goodreads.user_info(@user_id)
  @all_reviews = Goodreads.all_reviews(@user_id)
  @by_day = Goodreads.reviews_by_date(@all_reviews)
  @by_month = Goodreads.reviews_by_date(@all_reviews, fill=:month)

  @by_month = Hash[@by_month.
                   group_by {|d, r| d.strftime("%Y/%m")}.
                   map {|k, v| [k, Hash[v]]}]

  @page_title = "Goodreads timeline for #{@user_id}"
  @excanvas = '/ext/excanvas.min.js'
  @scripts = ['/ext/jquery-1.7.min.js', '/ext/flot.min.js',
              '/ext/qtip.min.js', '/ext/chart.js', '/ext/tooltip.js']

  haml(:timeline)
end
