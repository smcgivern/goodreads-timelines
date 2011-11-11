require 'lib/goodreads'
require 'sinatra'
require 'sinatra/reloader'

set(:views, Proc.new {File.join(root, 'view')})

helpers do
  def partial(template, variables={})
    haml(template, {:layout => false}, variables)
  end
end

get('/ext/style.css') {scss(:style)}

get('/::user_id/?') do
  @user_id = params['user_id']
  @all_reviews = Goodreads.all_reviews(@user_id)
  @by_date = Goodreads.reviews_by_date(@all_reviews, fill=:month)

  @by_month = Hash[@by_date.
                   group_by {|d, r| d.strftime("%Y/%m")}.
                   map {|k, v| [k, Hash[v]]}]

  haml(:timeline)
end
