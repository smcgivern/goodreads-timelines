require 'date'
require 'fileutils'
require 'open-uri'

require 'addressable/template'
require 'nokogiri'

module Goodreads
  OPTIONS = {
    :between_requests => 1,
    :page_size => 200,
    :cache_for => 24 * 60 * 60,
    :cache_dir => File.join(File.dirname(__FILE__), '../cache'),
  }

  API_KEY = open('goodreads.key').read.strip
  DATE_FORMAT = '%a %b %d %H:%M:%S %Z %Y'

  ADDRESSES = {
    :review => {
      :list =>
      ::Addressable::Template.new('http://www.goodreads.com/review/list/{user_id}.xml?key={api_key}&page={page}&per_page={page_size}&sort=date_read&order=d&shelf=read&v=2'),
    },
  }

  def self.ago(s); Time.now - s; end

  def self.limit_rate(meth, options=OPTIONS)
    @@last_call ||= {}
    @@last_call[meth] ||= ago(2 * options[:between_requests])

    sleep(1) while @@last_call[meth] > ago(options[:between_requests])

    @@last_call[meth] = Time.now
  end

  def self.cache_to(filename, options=OPTIONS)
    if options[:cache_dir]
      FileUtils.mkdir_p(options[:cache_dir])

      cache_file = File.join(options[:cache_dir], filename)

      if File.exist?(cache_file)
        if File.mtime(cache_file) > ago(options[:cache_for])
          return open(cache_file)
        else
          File.delete(cache_file)
        end
      end
    end

    block_return = yield

    open(cache_file, 'w').puts(block_return) if options[:cache_dir]

    block_return
  end

  # Picks the page number of user_id's reviews using the Goodreads
  # API.
  def self.list_page(user_id, page=1, options=OPTIONS)
    review_list = cache_to("review_list_#{user_id}.xml", options) do
      expansions = {
        'user_id' => user_id,
        'api_key' => API_KEY,
        'page' => page,
        'page_size' => options[:page_size],
      }

      limit_rate(:list_page, options)
      open(ADDRESSES[:review][:list].expand(expansions)).read
    end

    Nokogiri::XML(review_list)
  end

  # Gets all reviews for the specified user ID, using #list_page. If
  # there's more than one page, adds them all to the array.
  def self.all_reviews(user_id, options=OPTIONS)
    reviews = list_page(user_id).at('reviews')

    # Are there more books than fit on this page?
    if reviews['end'] != reviews['total']
      pages = (reviews['total'].to_i / options[:page_size].to_f).ceil

      2.upto(pages) {|i| reviews << list_page(user_id, i).at('reviews')}
    end

    reviews
  end

  # Puts reviews in a hash with each key representing the read
  # date. If a book hasn't been marked as read, it gets dropped from
  # the output.
  #
  # If fill == :day, then all days between the first and last review
  # will be included, even if they have no reviews.
  #
  # If fill == :month, then the same as the above applies but the
  # endpoints are the first day of the first month and the last day of
  # the last month in the set.
  def self.reviews_by_date(reviews, fill=false)
    by_date = {}

    reviews.search('review').each do |review|
      next if review.at('read_at').inner_text.empty?

      date = Date.strptime(review.at('read_at').inner_text, DATE_FORMAT)

      by_date[date] ||= []
      by_date[date] << review
    end

    if fill
      start_date = by_date.keys.min
      end_date = by_date.keys.max

      if fill == :month
        start_date = Date.civil(start_date.year, start_date.month, 1)
        end_date = Date.civil(end_date.year, end_date.month, -1)
      end

      start_date.upto(end_date) {|date| by_date[date] ||= []}
    end

    by_date
  end
end
